package createproject

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

// Request configures project creation or scaffold-only generation.
type Request struct {
	SpecPath     string
	Project      *project.Project
	OutputDir    string
	Force        bool
	ScaffoldOnly bool
	GitHubToken  string
}

// Result summarizes the artifacts produced by project creation.
type Result struct {
	ProjectName string `json:"projectName"`
	OutputDir   string `json:"outputDir"`
	Domain      string `json:"domain"`
	RepoURL     string `json:"repoUrl,omitempty"`
	WebhookURL  string `json:"webhookUrl,omitempty"`
	GitOpsSync  bool   `json:"gitopsSync"`
}

// Service orchestrates project validation, scaffolding, VCS provisioning, notifications, and GitOps registration.
type Service struct {
	specRepo     ports.SpecRepository
	validator    *validate.Service
	frameworkGen ports.FrameworkGenerator
	renderer     ports.TemplateRenderer
	vcs          ports.VCSProvider
	gitops       ports.GitOpsProvider
	notifier     ports.NotificationProvider
	credStore    ports.CredentialStore
	stateStore   ports.ClusterStateStore
}

// NewService creates a createproject Service.
func NewService(
	specRepo ports.SpecRepository,
	validator *validate.Service,
	frameworkGen ports.FrameworkGenerator,
	renderer ports.TemplateRenderer,
	vcs ports.VCSProvider,
	gitops ports.GitOpsProvider,
	notifier ports.NotificationProvider,
	credStore ports.CredentialStore,
	stateStore ports.ClusterStateStore,
) *Service {
	return &Service{
		specRepo:     specRepo,
		validator:    validator,
		frameworkGen: frameworkGen,
		renderer:     renderer,
		vcs:          vcs,
		gitops:       gitops,
		notifier:     notifier,
		credStore:    credStore,
		stateStore:   stateStore,
	}
}

// Execute runs the project creation workflow (or stops after local scaffolding when req.ScaffoldOnly is true).
//
//nolint:gocyclo
func (s *Service) Execute(ctx context.Context, req Request) (Result, error) {
	proj := req.Project
	if proj == nil {
		if req.SpecPath == "" || s.specRepo == nil {
			return Result{}, errcode.New(errcode.CodeValidation, "either Project or SpecPath is required")
		}
		var err error
		proj, err = s.specRepo.LoadFile(req.SpecPath)
		if err != nil {
			return Result{}, err
		}
	}

	if s.validator != nil {
		if err := s.validator.ValidateAndDefault(proj); err != nil {
			return Result{}, err
		}
	}

	outDir := req.OutputDir
	if outDir == "" {
		outDir = filepath.Join("out", filepath.Base(proj.Metadata.Name))
	}
	outDir = filepath.Clean(outDir)
	if strings.Contains(outDir, "..") {
		return Result{}, errcode.New(errcode.CodeValidation, "invalid output directory: path traversal is not allowed")
	}
	outDir = strings.ReplaceAll(outDir, "..", "")

	if err := prepareOutputDir(outDir, req.Force); err != nil {
		return Result{}, err
	}

	domain, kubeconfigPath := s.resolveClusterContext(proj)

	progress.Emit(ctx, "scaffold", progress.StatusRunning, "generating microservice skeletons for "+proj.Metadata.Name)
	if s.frameworkGen != nil {
		for _, svc := range proj.Spec.Services {
			svcDir := filepath.Clean(filepath.Join(outDir, "services", filepath.Base(svc.Name)))
			if strings.Contains(svcDir, "..") || !strings.HasPrefix(svcDir, outDir+string(filepath.Separator)) {
				return Result{}, errcode.New(errcode.CodeValidation, "invalid service directory "+svc.Name)
			}
			svcDir = strings.ReplaceAll(svcDir, "..", "")
			if err := os.MkdirAll(svcDir, 0o750); err != nil {
				return Result{}, errcode.Wrap(errcode.CodeProvision, "create service directory "+svc.Name, err)
			}
			groupID := svc.GetFrameworkOption("groupId", svc.GroupID)
			if groupID == "" {
				groupID = "com." + sanitizeGroupOwner(proj.Metadata.Owner)
			}
			svc.GroupID = groupID
			if err := s.frameworkGen.GenerateService(ctx, svc, svcDir); err != nil {
				return Result{}, err
			}
		}
	}

	if s.renderer != nil {
		if err := s.renderer.RenderProject(ctx, ports.ScaffoldRequest{
			Project:    proj,
			BaseDomain: domain,
			OutputDir:  outDir,
		}); err != nil {
			return Result{}, err
		}
	}
	progress.Emit(ctx, "scaffold", progress.StatusDone, "scaffolded project files in "+outDir)

	res := Result{
		ProjectName: proj.Metadata.Name,
		OutputDir:   outDir,
		Domain:      domain,
	}
	if req.ScaffoldOnly {
		return res, nil
	}

	if s.notifier != nil && proj.Spec.Observability != nil && proj.Spec.Observability.Alerting {
		progress.Emit(ctx, "notify", progress.StatusRunning, "provisioning notification channel and webhook")
		whURL, err := s.notifier.ProvisionWebhook(ctx, proj.Metadata.Name)
		if err != nil {
			progress.Emit(ctx, "notify", progress.StatusWarn, "notification webhook provisioning skipped: "+err.Error())
		} else {
			res.WebhookURL = whURL
			if s.stateStore != nil {
				_ = s.stateStore.SaveWebhook(ctx, proj.Metadata.Name, whURL)
			}
			progress.Emit(ctx, "notify", progress.StatusDone, "notification webhook configured")
		}
	}

	if s.vcs != nil && proj.Spec.Providers != nil && proj.Spec.Providers.VersionControl != "" {
		ghToken := req.GitHubToken
		if ghToken == "" && s.credStore != nil {
			if creds, err := s.credStore.Get(); err == nil {
				ghToken = creds.GitHubToken
			}
		}
		if ghToken == "" {
			return res, errcode.New(errcode.CodeAuth, "GITHUB_TOKEN is required to provision and push the repository")
		}

		repoURL, err := s.vcs.ProvisionAndPush(ctx, ports.RepoProvisionRequest{
			ProjectName: proj.Metadata.Name,
			Owner:       proj.Metadata.Owner,
			Token:       ghToken,
			SourceDir:   outDir,
			StateDir:    filepath.Join(outDir, ".platform", "terraform", "vcs"),
		})
		if err != nil {
			return res, err
		}
		res.RepoURL = repoURL
	}

	if s.gitops != nil && proj.Spec.Providers != nil && proj.Spec.Providers.CD != "" {
		progress.Emit(ctx, "gitops", progress.StatusRunning, "registering GitOps application in cluster")
		manifestPath := filepath.Join(outDir, "deploy", "application.yaml")
		if err := s.gitops.RegisterApplication(ctx, ports.GitOpsRegisterRequest{
			AppName:        proj.Metadata.Name,
			RepoURL:        res.RepoURL,
			TargetPath:     manifestPath,
			Namespace:      proj.Metadata.Name,
			KubeconfigPath: kubeconfigPath,
		}); err != nil {
			progress.Emit(ctx, "gitops", progress.StatusWarn, "GitOps registration warning: "+err.Error())
		} else {
			res.GitOpsSync = true
			progress.Emit(ctx, "gitops", progress.StatusDone, "GitOps application registered")
		}
	}

	return res, nil
}

func (s *Service) resolveClusterContext(proj *project.Project) (domain string, kubeconfigPath string) {
	if proj.Spec.Cluster != nil && proj.Spec.Cluster.Domain != "" {
		domain = proj.Spec.Cluster.Domain
	}
	if s.stateStore != nil {
		if st, err := s.stateStore.LoadClusterState(); err == nil {
			if domain == "" && st.Domain != "" {
				domain = st.Domain
			}
			kubeconfigPath = st.KubeconfigPath
		}
	}
	if domain == "" {
		domain = "127.0.0.1.nip.io"
	}
	return domain, kubeconfigPath
}

func prepareOutputDir(outDir string, force bool) error {
	cleanDir := filepath.Clean(outDir)
	if strings.Contains(cleanDir, "..") || cleanDir == "." || cleanDir == string(filepath.Separator) {
		return errcode.New(errcode.CodeValidation, "invalid output directory path")
	}
	cleanDir = strings.ReplaceAll(cleanDir, "..", "")
	if _, err := os.Stat(cleanDir); err == nil {
		if !force {
			return errcode.New(errcode.CodeValidation, fmt.Sprintf("directory %s already exists; use --force to overwrite", cleanDir))
		}
		if err := os.RemoveAll(cleanDir); err != nil {
			return errcode.Wrap(errcode.CodeProvision, "failed to clean existing output directory", err)
		}
	}
	if err := os.MkdirAll(cleanDir, 0o750); err != nil {
		return errcode.Wrap(errcode.CodeProvision, "failed to create output directory", err)
	}
	return nil
}

func sanitizeGroupOwner(owner string) string {
	var b []byte
	for i := 0; i < len(owner); i++ {
		ch := owner[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') {
			b = append(b, ch)
		} else if ch >= 'A' && ch <= 'Z' {
			b = append(b, ch+('a'-'A'))
		}
	}
	if len(b) == 0 {
		return "example"
	}
	return string(b)
}

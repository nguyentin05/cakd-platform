package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	cataloguc "github.com/nguyentin05/cakd-platform/internal/usecase/catalog"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

// Services contains the application services required by the interactive TUI wizard.
type Services struct {
	ValidateUC *validate.Service
	SpecRepo   ports.SpecRepository
	CreateUC   *createproject.Service
	CatalogUC  *cataloguc.Service
	CredStore  ports.CredentialStore
}

// Options configures the interactive TUI project creation wizard.
type Options struct {
	OutputPath string
	Force      bool
}

// RunWizard runs an interactive step-by-step terminal wizard that builds a canonical
// platform.yaml without requiring manual YAML authoring, and optionally executes project bootstrap.
func RunWizard(ctx context.Context, in io.Reader, out io.Writer, s *Services, opts Options) error {
	scanner := bufio.NewScanner(in)
	outPath := opts.OutputPath
	if outPath == "" {
		outPath = "platform.yaml"
	}

	_, _ = fmt.Fprintln(out, "╭────────────────────────────────────────────────────────╮")
	_, _ = fmt.Fprintln(out, "│      CAKD Platform — Interactive Project Wizard        │")
	_, _ = fmt.Fprintln(out, "╰────────────────────────────────────────────────────────╯")

	vcsOptions := resolveVCSOptions(ctx, s)
	projName := askDefault(scanner, out, "1. Project name", "my-platform")
	vcsProvider := askChoice(scanner, out, "2. Version control provider", vcsOptions, catalog.Defaults.VersionControl)
	ownerLabel := "3. GitHub owner / organization"
	if vcsProvider != catalog.ProviderGitHub {
		ownerLabel = "3. Repository owner"
	}
	owner := askDefault(scanner, out, ownerLabel, "my-org")
	provider := askChoice(scanner, out, "4. Cluster provider", catalog.ClusterProviders, catalog.Defaults.ClusterProvider)
	clusterSpec := promptClusterSpec(scanner, out, provider)

	svcName := askDefault(scanner, out, "5. Primary microservice name", "api-service")
	javaVer := askChoice(scanner, out, "6. Java version", catalog.ValidJavaVersions, catalog.Defaults.LanguageVersion[catalog.LangJavaSpringBoot])
	buildTool := askChoice(scanner, out, "7. Build tool", []string{"maven-project", "gradle-project"}, "maven-project")

	backing, uses := promptBackingSpec(scanner, out, svcName)

	proj := &project.Project{
		APIVersion: project.CanonicalAPIVersion,
		Kind:       project.KindProject,
		Metadata: project.Metadata{
			Name:        projName,
			Owner:       owner,
			Environment: catalog.Defaults.Environment,
		},
		Spec: project.Spec{
			Cluster: clusterSpec,
			Providers: &project.ProvidersSpec{
				VersionControl: vcsProvider,
				CI:             catalog.Defaults.CI,
				CD:             catalog.Defaults.CD,
			},
			Backing: backing,
			Services: []project.ServiceSpec{
				{
					Name:         svcName,
					Language:     catalog.LangJavaSpringBoot,
					Version:      javaVer,
					ProjectBuild: buildTool,
					Uses:         uses,
				},
			},
		},
	}

	if err := s.ValidateUC.ValidateAndDefault(proj); err != nil {
		return err
	}

	yamlBytes, err := s.SpecRepo.MarshalCanonical(proj)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(out, "\n── Generated platform.yaml Preview ──────────────────────")
	_, _ = fmt.Fprint(out, string(yamlBytes))
	_, _ = fmt.Fprintln(out, "─────────────────────────────────────────────────────────")

	if err := s.SpecRepo.SaveFile(outPath, proj); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "✔ Saved configuration to %s\n", outPath)

	return promptAndRunCreate(ctx, scanner, out, s, proj, vcsProvider, opts.Force)
}

func resolveVCSOptions(ctx context.Context, s *Services) []string {
	vcsOptions := []string{catalog.ProviderGitHub, catalog.ProviderLocalGit}
	if s.CatalogUC != nil {
		if provs, err := s.CatalogUC.ListProviders(ctx, "vcs"); err == nil && len(provs) > 0 {
			ids := make([]string, 0, len(provs))
			for _, p := range provs {
				ids = append(ids, p.ID)
			}
			if len(ids) > 0 {
				return ids
			}
		}
	}
	return vcsOptions
}

func promptClusterSpec(scanner *bufio.Scanner, out io.Writer, provider string) *project.ClusterSpec {
	spec := &project.ClusterSpec{
		Provider: provider,
	}
	if provider == catalog.ClusterAWSEC2 {
		spec.Region = askDefault(scanner, out, "   AWS Region", catalog.Defaults.Region)
		spec.InstanceType = askChoice(scanner, out, "   EC2 Instance Type", catalog.InstanceTypes, catalog.Defaults.InstanceType)
		spec.DiskSize = catalog.Defaults.DiskSize
	}
	return spec
}

func promptBackingSpec(scanner *bufio.Scanner, out io.Writer, svcName string) ([]project.BackingSpec, []string) {
	if !askYesNo(scanner, out, "8. Provision PostgreSQL backing database? [y/N]", false) {
		return nil, nil
	}
	dbName := askDefault(scanner, out, "   PostgreSQL resource name", svcName+"-db")
	backing := []project.BackingSpec{
		{
			Name:    dbName,
			Type:    catalog.BackingPostgreSQL,
			Version: catalog.Defaults.DBVersion[catalog.BackingPostgreSQL],
			Storage: catalog.Defaults.Storage,
		},
	}
	return backing, []string{dbName}
}

func promptAndRunCreate(ctx context.Context, scanner *bufio.Scanner, out io.Writer, s *Services, proj *project.Project, vcsProvider string, force bool) error {
	if !askYesNo(scanner, out, "9. Run 'cakd-cli create' now? [y/N]", false) {
		return nil
	}
	ctxWithSink := progress.WithSink(ctx, progress.SinkFunc(func(ev progress.Event) {
		_, _ = fmt.Fprintf(out, "  [%s] (%s) %s\n", ev.Phase, ev.Status, ev.Message)
	}))
	var vcsToken string
	if s.CredStore != nil && vcsProvider == catalog.ProviderGitHub {
		if creds, err := s.CredStore.Get(); err == nil {
			vcsToken = creds.GitHubToken
		}
	}
	res, err := s.CreateUC.Execute(ctxWithSink, createproject.Request{
		Project:  proj,
		Force:    force,
		VCSToken: vcsToken,
	})
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "✔ Project created at %s (domain: %s)\n", res.OutputDir, res.Domain)
	return nil
}

func askDefault(scanner *bufio.Scanner, out io.Writer, label, defaultVal string) string {
	_, _ = fmt.Fprintf(out, "%s [%s]: ", label, defaultVal)
	if scanner.Scan() {
		if val := strings.TrimSpace(scanner.Text()); val != "" {
			return val
		}
	}
	return defaultVal
}

func askChoice(scanner *bufio.Scanner, out io.Writer, label string, options []string, defaultVal string) string {
	_, _ = fmt.Fprintf(out, "%s (%s) [%s]: ", label, strings.Join(options, "/"), defaultVal)
	if scanner.Scan() {
		if val := strings.TrimSpace(scanner.Text()); val != "" {
			return val
		}
	}
	return defaultVal
}

func askYesNo(scanner *bufio.Scanner, out io.Writer, label string, defaultYes bool) bool {
	_, _ = fmt.Fprintf(out, "%s: ", label)
	if scanner.Scan() {
		val := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if val == "y" || val == "yes" {
			return true
		}
		if val == "n" || val == "no" {
			return false
		}
	}
	return defaultYes
}

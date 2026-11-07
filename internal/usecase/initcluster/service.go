package initcluster

import (
	"context"
	"path/filepath"

	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

// Request configures cluster initialization from CLI flags, MCP tools, or a platform.yaml file.
type Request struct {
	SpecPath       string
	Project        *project.Project
	ArgoCD         bool
	Monitoring     bool
	Logging        bool
	GeminiAPIKey   string
	DiscordWebhook string
}

// Service orchestrates cloud host provisioning (when aws-ec2) and K3s/Helm cluster bootstrapping.
type Service struct {
	specRepo     ports.SpecRepository
	validator    *validate.Service
	hostProv     ports.CloudHostProvisioner
	bootstrapper ports.ClusterBootstrapper
	stateStore   ports.ClusterStateStore
}

// NewService creates an initcluster Service.
func NewService(
	specRepo ports.SpecRepository,
	validator *validate.Service,
	hostProv ports.CloudHostProvisioner,
	bootstrapper ports.ClusterBootstrapper,
	stateStore ports.ClusterStateStore,
) *Service {
	return &Service{
		specRepo:     specRepo,
		validator:    validator,
		hostProv:     hostProv,
		bootstrapper: bootstrapper,
		stateStore:   stateStore,
	}
}

// Execute provisions a cloud host if needed, bootstraps the cluster add-ons, and persists ClusterState.
//
//nolint:gocyclo
func (s *Service) Execute(ctx context.Context, req Request) (ports.ClusterState, error) {
	proj := req.Project
	if proj == nil && req.SpecPath != "" && s.specRepo != nil {
		loaded, err := s.specRepo.LoadFile(req.SpecPath)
		if err != nil {
			return ports.ClusterState{}, err
		}
		if s.validator != nil {
			if err := s.validator.ValidateAndDefault(loaded); err != nil {
				return ports.ClusterState{}, err
			}
		}
		proj = loaded
	}

	opts := ports.ClusterBootstrapOptions{
		Provider:       catalog.ClusterLocal,
		InstallArgoCD:  req.ArgoCD,
		InstallMonitor: req.Monitoring,
		InstallLogging: req.Logging,
		InstallAgent:   req.Monitoring,
		GeminiAPIKey:   req.GeminiAPIKey,
		DiscordWebhook: req.DiscordWebhook,
	}

	if proj != nil && proj.Spec.Cluster != nil {
		c := proj.Spec.Cluster
		if c.Provider != "" {
			opts.Provider = c.Provider
		}
		opts.Host = c.Host
		opts.SSHKeyPath = c.SSHKeyPath
		opts.SSHUser = c.SSHUser
		opts.Domain = c.Domain

		if c.Addons != nil {
			if c.Addons.ArgoCD != nil {
				opts.InstallArgoCD = *c.Addons.ArgoCD
			}
			if c.Addons.Monitoring != nil {
				opts.InstallMonitor = *c.Addons.Monitoring
			}
			if c.Addons.Logging != nil {
				opts.InstallLogging = *c.Addons.Logging
			}
			if c.Addons.Agent != nil {
				opts.InstallAgent = *c.Addons.Agent
			}
		}

		if c.Provider == catalog.ClusterAWSEC2 && opts.Host == "" && s.hostProv != nil {
			if c.InstanceType == "t3.micro" {
				progress.Emit(ctx, "cluster-check", progress.StatusWarn, "t3.micro (1GB RAM) is tight for K3s + ArgoCD + Prometheus; t3.small or t3.medium is recommended")
			}
			res, err := s.hostProv.ProvisionHost(ctx, ports.CloudHostProvisionRequest{
				ProjectName:  proj.Metadata.Name,
				Region:       c.Region,
				InstanceType: c.InstanceType,
				DiskSize:     c.DiskSize,
				WorkDir:      filepath.Join("out", proj.Metadata.Name, ".platform", "terraform", "cluster"),
			})
			if err != nil {
				return ports.ClusterState{}, err
			}
			opts.Host = res.PublicIP
			opts.SSHKeyPath = res.SSHKeyPath
			if opts.Domain == "" {
				opts.Domain = res.PublicIP + ".nip.io"
			}
		}
	}

	state, err := s.bootstrapper.BootstrapCluster(ctx, opts)
	if err != nil {
		return ports.ClusterState{}, err
	}

	if s.stateStore != nil {
		_ = s.stateStore.SaveClusterState(state)
	}
	return state, nil
}

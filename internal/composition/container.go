package composition

import (
	"context"
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/domain/catalog"
	"github.com/nguyentin05/cakd-platform/internal/domain/errcode"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/catalogstore"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/configfile"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/credstore"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/k8s"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/argocd"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/discord"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/gemini"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/github"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/localgit"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/loki"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/prometheus"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/springboot"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/scaffold"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/statestore"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/terraform"
	"github.com/nguyentin05/cakd-platform/internal/usecase/agenttools"
	"github.com/nguyentin05/cakd-platform/internal/usecase/alertdispatch"
	"github.com/nguyentin05/cakd-platform/internal/usecase/auth"
	cataloguc "github.com/nguyentin05/cakd-platform/internal/usecase/catalog"
	"github.com/nguyentin05/cakd-platform/internal/usecase/createproject"
	"github.com/nguyentin05/cakd-platform/internal/usecase/initcluster"
	"github.com/nguyentin05/cakd-platform/internal/usecase/observe"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
	"github.com/nguyentin05/cakd-platform/internal/usecase/validate"
)

// Container wires all outbound adapters and inbound application use cases.
type Container struct {
	SpecRepo   ports.SpecRepository
	CredStore  ports.CredentialStore
	StateStore ports.ClusterStateStore

	ValidateUC *validate.Service
	InitUC     *initcluster.Service
	CreateUC   *createproject.Service
	ObserveUC  *observe.Service
	AuthUC     *auth.Service
	AgentTools *agenttools.Catalog
	CatalogUC  *cataloguc.Service
}

// NewContainer builds the production dependency injection container for CLI, TUI, and MCP binaries.
func NewContainer() *Container {
	specRepo := configfile.NewLoader()
	credStore := credstore.NewStore()
	stateStore := statestore.NewStore()
	tfProv := terraform.NewProvisioner()
	k8sBootstrap := k8s.NewBootstrapper(credStore)
	renderer := scaffold.NewRenderer()
	springGen := springboot.NewGenerator()
	ghVCS := github.NewClient(tfProv)
	localGitVCS := localgit.NewClient()
	vcsRouter := NewVCSRouter(map[string]ports.VCSProvider{
		catalog.ProviderGitHub:   ghVCS,
		catalog.ProviderLocalGit: localGitVCS,
		"local":                  localGitVCS,
	})
	argoCD := argocd.NewClient()
	promClient := prometheus.NewClient()
	lokiClient := loki.NewClient()

	discClient := discord.NewClientWithStore(credStore)

	valUC := validate.NewService()
	initUC := initcluster.NewService(specRepo, valUC, tfProv, k8sBootstrap, stateStore)
	createUC := createproject.NewService(specRepo, valUC, springGen, renderer, vcsRouter, argoCD, discClient, credStore, stateStore)
	obsUC := observe.NewService(
		specRepo,
		promClient,
		lokiClient,
		func(apiKey, model string) ports.AIAnalyzer {
			return gemini.NewClient(apiKey, model)
		},
		discClient,
		credStore,
		stateStore,
	)
	authUC := auth.NewService(credStore)
	tools := agenttools.NewCatalog(specRepo, valUC, createUC, initUC, obsUC)
	catalogStore := catalogstore.NewStore()
	catalogUC := cataloguc.NewService(catalogStore)

	return &Container{
		SpecRepo:   specRepo,
		CredStore:  credStore,
		StateStore: stateStore,
		ValidateUC: valUC,
		InitUC:     initUC,
		CreateUC:   createUC,
		ObserveUC:  obsUC,
		AuthUC:     authUC,
		AgentTools: tools,
		CatalogUC:  catalogUC,
	}
}

// ToServices converts the Container to the decoupled cmdutil.Services bundle.
func (c *Container) ToServices() *cmdutil.Services {
	return &cmdutil.Services{
		ValidateUC: c.ValidateUC,
		InitUC:     c.InitUC,
		CreateUC:   c.CreateUC,
		ObserveUC:  c.ObserveUC,
		AuthUC:     c.AuthUC,
		AgentTools: c.AgentTools,
		CatalogUC:  c.CatalogUC,
		SpecRepo:   c.SpecRepo,
		CredStore:  c.CredStore,
		StateStore: c.StateStore,
	}
}

// NewAlertDispatcher constructs the Alertmanager webhook dispatcher for the cakd-agent daemon.
func NewAlertDispatcher(defaultWebhookURL, discordToken, discordGuildID, geminiKey, geminiModel string) *alertdispatch.Service {
	stateStore := statestore.NewStore()
	notifier := discord.NewClient(discordToken, discordGuildID)
	var analyzer ports.AIAnalyzer
	if geminiKey != "" {
		analyzer = gemini.NewClient(geminiKey, geminiModel)
	}
	return alertdispatch.NewService(defaultWebhookURL, notifier, analyzer, stateStore)
}

// VCSRouter dispatches VCS operations to provider-specific implementations based on req.Provider.
type VCSRouter struct {
	providers map[string]ports.VCSProvider
}

// NewVCSRouter creates a new VCSRouter.
func NewVCSRouter(providers map[string]ports.VCSProvider) *VCSRouter {
	return &VCSRouter{providers: providers}
}

var _ ports.VCSProvider = (*VCSRouter)(nil)

// ProvisionAndPush routes repository provisioning and push to the target provider.
func (r *VCSRouter) ProvisionAndPush(ctx context.Context, req ports.RepoProvisionRequest) (string, error) {
	providerKey := req.Provider
	if providerKey == "" {
		providerKey = catalog.ProviderGitHub
	}
	p, ok := r.providers[providerKey]
	if !ok || p == nil {
		p = r.providers[catalog.ProviderGitHub]
	}
	if p == nil {
		return "", errcode.New(errcode.CodeProvision, fmt.Sprintf("unsupported vcs provider: %s", req.Provider))
	}
	return p.ProvisionAndPush(ctx, req)
}

// Rollback routes repository cleanup to the target provider.
func (r *VCSRouter) Rollback(ctx context.Context, req ports.RepoProvisionRequest) error {
	providerKey := req.Provider
	if providerKey == "" {
		providerKey = catalog.ProviderGitHub
	}
	p, ok := r.providers[providerKey]
	if !ok || p == nil {
		p = r.providers[catalog.ProviderGitHub]
	}
	if p == nil {
		return nil
	}
	return p.Rollback(ctx, req)
}

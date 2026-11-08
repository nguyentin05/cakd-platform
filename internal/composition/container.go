package composition

import (
	"github.com/nguyentin05/cakd-platform/internal/adapter/cli/cmdutil"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/configfile"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/credstore"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/k8s"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/argocd"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/discord"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/gemini"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/github"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/loki"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/prometheus"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/providers/springboot"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/scaffold"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/statestore"
	"github.com/nguyentin05/cakd-platform/internal/infrastructure/terraform"
	"github.com/nguyentin05/cakd-platform/internal/usecase/agenttools"
	"github.com/nguyentin05/cakd-platform/internal/usecase/alertdispatch"
	"github.com/nguyentin05/cakd-platform/internal/usecase/auth"
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
	argoCD := argocd.NewClient()
	promClient := prometheus.NewClient()
	lokiClient := loki.NewClient()

	discClient := discord.NewClientWithStore(credStore)

	valUC := validate.NewService()
	initUC := initcluster.NewService(specRepo, valUC, tfProv, k8sBootstrap, stateStore)
	createUC := createproject.NewService(specRepo, valUC, springGen, renderer, ghVCS, argoCD, discClient, credStore, stateStore)
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

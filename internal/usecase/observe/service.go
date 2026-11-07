package observe

import (
	"context"
	"fmt"

	"github.com/nguyentin05/cakd-platform/internal/domain/diagnostic"
	"github.com/nguyentin05/cakd-platform/internal/domain/project"
	"github.com/nguyentin05/cakd-platform/internal/platform/progress"
	"github.com/nguyentin05/cakd-platform/internal/usecase/ports"
)

// Request configures an on-demand workload diagnostic run.
type Request struct {
	SpecPath     string
	Project      *project.Project
	Namespace    string
	SendToNotify bool
}

// Service collects metrics and logs and performs AI root-cause analysis.
type Service struct {
	specRepo   ports.SpecRepository
	metrics    ports.MetricsCollector
	logs       ports.LogCollector
	aiFactory  func(apiKey, model string) ports.AIAnalyzer
	notifier   ports.NotificationProvider
	credStore  ports.CredentialStore
	stateStore ports.ClusterStateStore
}

// NewService creates an observe Service.
func NewService(
	specRepo ports.SpecRepository,
	metrics ports.MetricsCollector,
	logs ports.LogCollector,
	aiFactory func(apiKey, model string) ports.AIAnalyzer,
	notifier ports.NotificationProvider,
	credStore ports.CredentialStore,
	stateStore ports.ClusterStateStore,
) *Service {
	return &Service{
		specRepo:   specRepo,
		metrics:    metrics,
		logs:       logs,
		aiFactory:  aiFactory,
		notifier:   notifier,
		credStore:  credStore,
		stateStore: stateStore,
	}
}

// Execute collects cluster observability data and returns a diagnostic Report.
//
//nolint:gocyclo
func (s *Service) Execute(ctx context.Context, req Request) (diagnostic.Report, error) {
	proj := req.Project
	if proj == nil && req.SpecPath != "" && s.specRepo != nil {
		loaded, err := s.specRepo.LoadFile(req.SpecPath)
		if err != nil {
			return diagnostic.Report{}, err
		}
		proj = loaded
	}

	ns := req.Namespace
	if ns == "" && proj != nil {
		ns = proj.Metadata.Name
	}
	if ns == "" {
		ns = "default"
	}

	progress.Emit(ctx, "observe-metrics", progress.StatusRunning, "fetching Prometheus metrics for namespace "+ns)
	metricsSummary := "No metrics available."
	if s.metrics != nil {
		if m, err := s.metrics.CollectMetricsSummary(ctx, ns); err == nil && m != "" {
			metricsSummary = m
		} else if err != nil {
			progress.Emit(ctx, "observe-metrics", progress.StatusWarn, "metrics fetch warning: "+err.Error())
		}
	}

	progress.Emit(ctx, "observe-logs", progress.StatusRunning, "fetching Loki logs for namespace "+ns)
	logsSummary := "No logs available."
	if s.logs != nil {
		if l, err := s.logs.CollectLogsSummary(ctx, ns); err == nil && l != "" {
			logsSummary = l
		} else if err != nil {
			progress.Emit(ctx, "observe-logs", progress.StatusWarn, "logs fetch warning: "+err.Error())
		}
	}

	model := "gemini-2.5-flash"
	if proj != nil && proj.Spec.Observability != nil && proj.Spec.Observability.AI != nil && proj.Spec.Observability.AI.Model != "" {
		model = proj.Spec.Observability.AI.Model
	}
	var apiKey string
	if s.credStore != nil {
		if creds, err := s.credStore.Get(); err == nil {
			apiKey = creds.GeminiAPIKey
		}
	}

	prompt := fmt.Sprintf(`You are an expert DevOps engineer and Kubernetes administrator.
The user is experiencing issues with their application in the %q namespace.

--- METRICS ---
%s

--- LOGS ---
%s

Analyze this data:
1. Identify any crashing or restarting pods.
2. Determine the root cause from the logs.
3. Provide a concise, actionable fix in Markdown.`, ns, metricsSummary, logsSummary)

	progress.Emit(ctx, "observe-ai", progress.StatusRunning, "analyzing telemetry with AI model "+model)
	var analysis string
	if s.aiFactory != nil {
		analyzer := s.aiFactory(apiKey, model)
		var err error
		analysis, err = analyzer.Analyze(ctx, prompt)
		if err != nil {
			return diagnostic.Report{}, err
		}
	}
	progress.Emit(ctx, "observe-ai", progress.StatusDone, "AI root-cause analysis complete")

	report := diagnostic.Report{
		ProjectName:    ns,
		Namespace:      ns,
		MetricsSummary: metricsSummary,
		LogsSummary:    logsSummary,
		Analysis:       analysis,
	}

	if req.SendToNotify && s.notifier != nil && s.stateStore != nil {
		if hooks, err := s.stateStore.LoadWebhooks(); err == nil {
			if whURL := hooks[ns]; whURL != "" {
				if err := s.notifier.SendMessage(ctx, whURL, analysis); err == nil {
					report.Notified = true
				}
			}
		}
	}

	return report, nil
}

package observe_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/observe"
)

// MockMetricsFetcher implements monitoring.MetricsFetcher
type MockMetricsFetcher struct {
	FetchFunc func(namespace string) (string, error)
	Fetched   bool
}

func (m *MockMetricsFetcher) Fetch(namespace string) (string, error) {
	m.Fetched = true
	if m.FetchFunc != nil {
		return m.FetchFunc(namespace)
	}
	return "CPU=45%, Mem=120Mi", nil
}

// MockLogFetcher implements logging.LogFetcher
type MockLogFetcher struct {
	FetchFunc func(namespace string) (string, error)
	Fetched   bool
}

func (m *MockLogFetcher) Fetch(namespace string) (string, error) {
	m.Fetched = true
	if m.FetchFunc != nil {
		return m.FetchFunc(namespace)
	}
	return "ERROR: Out of memory", nil
}

// MockLLM implements llm.LLM
type MockLLM struct {
	AnalyzeFunc func(context string) (string, error)
	Analyzed    bool
}

func (m *MockLLM) Analyze(context string) (string, error) {
	m.Analyzed = true
	if m.AnalyzeFunc != nil {
		return m.AnalyzeFunc(context)
	}
	return "Mock Diagnosis: OOM Kill detected. Fix by increasing memory limit.", nil
}

func TestObserverService_Diagnose_Success(t *testing.T) {
	metrics := &MockMetricsFetcher{}
	logs := &MockLogFetcher{}
	ai := &MockLLM{}

	service := observe.NewObserverService(metrics, logs, ai)

	err := service.Diagnose("test-namespace")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !metrics.Fetched {
		t.Error("Expected metrics.Fetch to be called")
	}
	if !logs.Fetched {
		t.Error("Expected logs.Fetch to be called")
	}
	if !ai.Analyzed {
		t.Error("Expected ai.Analyze to be called")
	}
}

func TestObserverService_Diagnose_AIFailure(t *testing.T) {
	metrics := &MockMetricsFetcher{}
	logs := &MockLogFetcher{}
	ai := &MockLLM{
		AnalyzeFunc: func(context string) (string, error) {
			return "", errors.New("simulated AI error")
		},
	}

	service := observe.NewObserverService(metrics, logs, ai)

	err := service.Diagnose("test-namespace")
	if err == nil {
		t.Fatal("Expected error when AI analysis fails, but got nil")
	}

	if !strings.Contains(err.Error(), "ai analysis failed") {
		t.Errorf("Expected error to contain 'ai analysis failed', got: %v", err)
	}
}

package validate

import (
	"reflect"
	"testing"

	"github.com/nguyentin05/cakd-platform/internal/registry"
)

func TestGetValueByPath(t *testing.T) {
	type Observability struct {
		Alerting bool
	}
	type Providers struct {
		CI string
	}
	type TestConfig struct {
		Providers     Providers
		Observability *Observability
	}

	cfg := &TestConfig{
		Providers: Providers{
			CI: "github-actions",
		},
		Observability: &Observability{
			Alerting: true,
		},
	}
	v := reflect.ValueOf(cfg).Elem()

	t.Run("Valid Path Struct", func(t *testing.T) {
		val, ok := getValueByPath(v, "Providers.CI")
		if !ok || val.String() != "github-actions" {
			t.Errorf("Expected github-actions, got %v (ok: %v)", val, ok)
		}
	})

	t.Run("Valid Path Pointer", func(t *testing.T) {
		val, ok := getValueByPath(v, "Observability.Alerting")
		if !ok || !val.Bool() {
			t.Errorf("Expected true, got %v (ok: %v)", val, ok)
		}
	})

	t.Run("Invalid Path", func(t *testing.T) {
		_, ok := getValueByPath(v, "Providers.Unknown")
		if ok {
			t.Errorf("Expected false for unknown path")
		}
	})

	t.Run("Nil Pointer Path", func(t *testing.T) {
		cfg.Observability = nil
		_, ok := getValueByPath(v, "Observability.Alerting")
		if ok {
			t.Errorf("Expected false for nil pointer path")
		}
	})
}

func TestCheckCondition(t *testing.T) {
	t.Run("NotEmpty", func(t *testing.T) {
		if !checkCondition(reflect.ValueOf("test"), registry.NotEmpty) {
			t.Error("Expected NotEmpty to be true for 'test'")
		}
		if checkCondition(reflect.ValueOf(""), registry.NotEmpty) {
			t.Error("Expected NotEmpty to be false for ''")
		}
	})

	t.Run("IsTrue", func(t *testing.T) {
		if !checkCondition(reflect.ValueOf(true), registry.IsTrue) {
			t.Error("Expected IsTrue to be true for true")
		}
		if checkCondition(reflect.ValueOf(false), registry.IsTrue) {
			t.Error("Expected IsTrue to be false for false")
		}
	})

	t.Run("NotNil", func(t *testing.T) {
		type Dummy struct{}
		ptr := &Dummy{}
		if !checkCondition(reflect.ValueOf(ptr), registry.NotNil) {
			t.Error("Expected NotNil to be true for non-nil pointer")
		}
		var nilPtr *Dummy
		if checkCondition(reflect.ValueOf(nilPtr), registry.NotNil) {
			t.Error("Expected NotNil to be false for nil pointer")
		}
	})
}

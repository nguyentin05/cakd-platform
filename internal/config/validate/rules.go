package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/registry"
)

// evaluateRules processes the BusinessRules defined in the registry.
// It acts as a dynamic rule engine that interprets conditional paths
// (e.g., "Providers.CD") and enforces that their constraints are met.
// This is executed during Phase 3 of the parsing pipeline.
func evaluateRules(cfg any) error {
	v := reflect.ValueOf(cfg).Elem()

	for _, rule := range registry.BusinessRules {
		ifVal, ok := getValueByPath(v, rule.IfPath)
		if !ok {
			continue
		}

		if checkCondition(ifVal, rule.IfCond) {
			thenVal, ok := getValueByPath(v, rule.ThenPath)
			if !ok || !checkCondition(thenVal, rule.ThenCond) {
				return fmt.Errorf("%s", rule.ErrorMsg)
			}
		}
	}

	return nil
}

// getValueByPath is a reflection-based utility that extracts a nested field's value
// using a dot-notated string path (e.g., "Providers.CI").
// It handles pointer dereferencing automatically during traversal.
func getValueByPath(v reflect.Value, path string) (reflect.Value, bool) {
	parts := strings.Split(path, ".")
	curr := v
	for _, part := range parts {
		if curr.Kind() == reflect.Pointer {
			if curr.IsNil() {
				return reflect.Value{}, false
			}
			curr = curr.Elem()
		}
		if curr.Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		curr = curr.FieldByName(part)
		if !curr.IsValid() {
			return reflect.Value{}, false
		}
	}
	return curr, true
}

func checkCondition(v reflect.Value, cond registry.Condition) bool {
	switch cond {
	case registry.NotEmpty:
		return !v.IsZero()
	case registry.IsTrue:
		return v.Kind() == reflect.Bool && v.Bool()
	case registry.NotNil:
		return v.Kind() == reflect.Pointer && !v.IsNil()
	default:
		return false
	}
}

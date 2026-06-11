package defaults

import (
	"reflect"
)

// Apply injects implicit default values into the configuration struct.
// It recursively traverses the struct tree using reflection to resolve and evaluate
// `default` tags on empty fields.
//
// This operation is part of Phase 2 of the configuration parsing pipeline.
func Apply(cfg any) {
	dynamic(reflect.ValueOf(cfg).Elem())
}

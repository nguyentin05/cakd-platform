package defaults

import (
	"reflect"
)

// Apply injects implicit default values into the configuration struct.
// It uses reflection to dynamically traverse the struct tree and evaluate `default:""`
// tags on empty fields. It is executed during Phase 2 of the Parsing Pipeline, ensuring
// that logic validations in Phase 3 have a fully hydrated struct to work with.
func Apply(cfg any) {
	dynamic(reflect.ValueOf(cfg).Elem())
}

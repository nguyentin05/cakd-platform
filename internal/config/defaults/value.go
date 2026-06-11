package defaults

import (
	"reflect"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/registry"
)

// applyValue resolves and assigns default values to struct fields using tag definitions.
// It supports direct field resolution (e.g. "Providers.VersionControl") and dynamic map-based
// resolution (e.g. "map:LanguageVersion,key:Language") from the global registry defaults.
func applyValue(parentStruct reflect.Value, fieldVal reflect.Value, tag string) {
	defaultsVal := reflect.ValueOf(registry.Defaults)

	if strings.HasPrefix(tag, "map:") {
		parts := strings.Split(tag, ",")
		if len(parts) != 2 {
			return
		}
		mapName := strings.TrimPrefix(parts[0], "map:")
		keyFieldName := strings.TrimPrefix(parts[1], "key:")

		mapField := defaultsVal.FieldByName(mapName)
		if !mapField.IsValid() || mapField.Kind() != reflect.Map {
			return
		}

		keyField := parentStruct.FieldByName(keyFieldName)
		if !keyField.IsValid() {
			return
		}

		mapValue := mapField.MapIndex(keyField)
		if mapValue.IsValid() {
			fieldVal.Set(mapValue)
		}
	} else {
		parts := strings.Split(tag, ".")
		targetVal := defaultsVal
		for _, part := range parts {
			if targetVal.Kind() == reflect.Struct {
				targetVal = targetVal.FieldByName(part)
			} else {
				targetVal = reflect.Value{}
				break
			}
		}

		if targetVal.IsValid() && targetVal.Type().AssignableTo(fieldVal.Type()) {
			fieldVal.Set(targetVal)
		}
	}
}

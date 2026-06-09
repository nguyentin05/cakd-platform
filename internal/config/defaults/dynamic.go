package defaults

import (
	"reflect"
)

// dynamic is a recursive helper that traverses an arbitrary reflected value.
// It dives into pointers, slices, and arrays to locate structs. Once a struct
// is found, it delegates to dynamicStruct to inject default values.
func dynamic(v reflect.Value) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		dynamic(v.Elem())
		return
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			dynamic(v.Index(i))
		}
		return
	}

	if v.Kind() == reflect.Struct {
		dynamicStruct(v)
	}
}

// dynamicStruct iterates over all exported fields of a struct.
// It performs two critical operations:
//  1. Pointer Initialization: If a pointer field is nil but its underlying struct
//     contains fields with `default` tags, it instantiates the pointer to avoid nil panics.
//  2. Value Injection: If a field is at its zero-value and has a `default` tag,
//     it parses the tag and applies the appropriate default value.
func dynamicStruct(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		structField := v.Type().Field(i)
		if !structField.IsExported() {
			continue
		}

		if fieldVal.Kind() == reflect.Pointer && fieldVal.IsNil() {
			elemType := fieldVal.Type().Elem()
			if elemType.Kind() == reflect.Struct {
				hasDefault := false
				for j := 0; j < elemType.NumField(); j++ {
					if elemType.Field(j).Tag.Get("default") != "" {
						hasDefault = true
						break
					}
				}
				if hasDefault {
					fieldVal.Set(reflect.New(elemType))
				}
			}
		}

		if fieldVal.Kind() == reflect.Struct || fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Pointer {
			dynamic(fieldVal)
		}

		defaultTag := structField.Tag.Get("default")
		if defaultTag != "" && fieldVal.IsZero() {
			applyValue(v, fieldVal, defaultTag)
		}
	}
}

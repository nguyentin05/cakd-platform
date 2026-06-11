package defaults

import (
	"reflect"
)

// dynamic recursively inspects a reflected value, dereferencing pointers and
// iterating through slices/arrays to locate struct fields for default value injection.
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

// dynamicStruct iterates over all exported fields of the given struct value.
// It initializes nil pointers if their targets contain default tags, and
// resolves default values for empty fields.
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

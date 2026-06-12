package defaults

import (
	"reflect"
)

// dynamic recursively inspects a reflected value, dereferencing pointers and
// iterating through slices/arrays to locate struct fields for default value injection.
func dynamic(v reflect.Value) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return dynamic(v.Elem())
	}

	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			if err := dynamic(v.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}

	if v.Kind() == reflect.Struct {
		return dynamicStruct(v)
	}
	return nil
}

// dynamicStruct iterates over all exported fields of the given struct value.
// It initializes nil pointers if their targets contain default tags (auto-init).
// If a field is at its zero-value and has a `default` tag, it applies the default value.
// This side-effect is intentional to ensure "Convention over Configuration".
func dynamicStruct(v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		structField := v.Type().Field(i)
		if !structField.IsExported() {
			continue
		}

		if fieldVal.Kind() == reflect.Pointer && fieldVal.IsNil() {
			autoInitPointer(fieldVal)
		}

		if fieldVal.Kind() == reflect.Struct || fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Pointer {
			if err := dynamic(fieldVal); err != nil {
				return err
			}
		}

		defaultTag := structField.Tag.Get("default")
		if defaultTag != "" && fieldVal.IsZero() {
			if err := applyValue(v, fieldVal, defaultTag); err != nil {
				return err
			}
		}
	}
	return nil
}

// autoInitPointer automatically instantiates a nil pointer to a struct
// if that struct contains any fields with a `default` tag.
func autoInitPointer(fieldVal reflect.Value) {
	elemType := fieldVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return
	}
	for j := 0; j < elemType.NumField(); j++ {
		if elemType.Field(j).Tag.Get("default") != "" {
			fieldVal.Set(reflect.New(elemType))
			return
		}
	}
}

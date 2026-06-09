package validate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/nguyentin05/cakd-platform/internal/registry"
)

// Structure performs Phase 1 of the configuration parsing pipeline.
// It traverses the unmarshaled struct to ensure that all basic structural requirements
// are met (e.g., required tags are not zero-values, enum tags contain supported values,
// and parent configuration blocks are not entirely empty).
func Structure(cfg any) error {
	return tree(reflect.ValueOf(cfg), "config")
}

func tree(v reflect.Value, path string) error {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		if err := validateStructEmptyParent(v.Elem(), path); err != nil {
			return err
		}
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			if err := validateField(v.Field(i), v.Type().Field(i), path); err != nil {
				return err
			}
		}
	} else if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			if err := tree(v.Index(i), newPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateStructEmptyParent(elem reflect.Value, path string) error {
	if elem.Kind() != reflect.Struct {
		return nil
	}
	numFields := elem.NumField()
	if numFields == 0 {
		return nil
	}
	hasNonZero := false
	for i := 0; i < numFields; i++ {
		if !elem.Type().Field(i).IsExported() {
			continue
		}
		if !elem.Field(i).IsZero() {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		return fmt.Errorf("block '%s' is declared but empty, must define at least one child", path)
	}
	return nil
}

func validateField(fieldVal reflect.Value, structField reflect.StructField, path string) error {
	if !structField.IsExported() {
		return nil
	}
	fieldName := structField.Name
	yamlTag := structField.Tag.Get("yaml")
	if yamlTag != "" {
		tagName := strings.Split(yamlTag, ",")[0]
		if tagName != "" && tagName != "-" {
			fieldName = tagName
		}
	}

	newPath := fieldName
	if path != "" && path != "config" {
		newPath = path + "." + fieldName
	}

	if structField.Tag.Get("validate") == "required" && fieldVal.IsZero() {
		return fmt.Errorf("field '%s' is required", newPath)
	}

	enumTag := structField.Tag.Get("enum")
	if enumTag != "" && !fieldVal.IsZero() {
		if err := validateEnum(fieldVal, enumTag); err != nil {
			return err
		}
	}

	return tree(fieldVal, newPath)
}

func validateEnum(fieldVal reflect.Value, enumTag string) error {
	validValues, exists := registry.Enums[enumTag]
	if !exists {
		return nil
	}
	valStr := fmt.Sprintf("%v", fieldVal.Interface())
	for _, valid := range validValues {
		if valStr == valid {
			return nil
		}
	}
	return fmt.Errorf("unsupported %s: %q. allowed values are: %v", enumTag, valStr, validValues)
}

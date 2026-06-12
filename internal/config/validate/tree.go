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
	return tree(reflect.ValueOf(cfg), "")
}

// tree recursively traverses a reflect.Value representing the configuration tree
// to enforce structural rules and handle slice/array elements.
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

// validateStructEmptyParent verifies that if a struct pointer block is declared in the YAML
// (e.g. observability:), it contains at least one non-zero sub-field instead of being left empty.
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

// validateField checks an individual struct field's required tag constraints and enum limits,
// then continues recursive tree traversal on the field value.
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
	if path != "" {
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

// validateEnum checks if a field's value (or its elements if it is a slice/array)
// is registered as a valid option under registry.Enums.
func validateEnum(fieldVal reflect.Value, enumTag string) error {
	validValues, exists := registry.Enums[enumTag]
	if !exists {
		return nil
	}

	if fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Array {
		for i := 0; i < fieldVal.Len(); i++ {
			elem := fieldVal.Index(i)
			valStr := fmt.Sprintf("%v", elem.Interface())
			if err := checkEnum(valStr, validValues, enumTag); err != nil {
				return err
			}
		}
		return nil
	}

	valStr := fmt.Sprintf("%v", fieldVal.Interface())
	return checkEnum(valStr, validValues, enumTag)
}

// checkEnum evaluates if a single string matches any of the registered valid enum values.
func checkEnum(valStr string, validValues []string, enumTag string) error {
	for _, valid := range validValues {
		if valStr == valid {
			return nil
		}
	}
	return fmt.Errorf("unsupported %s: %q. allowed values are: %v", enumTag, valStr, validValues)
}

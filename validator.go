package ginji

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// validateStruct checks struct tags for validation rules.
// Supported tags: required, email, min=X, max=X
func validateStruct(v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil // Only validate structs
	}

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)
		tag := field.Tag.Get("ginji")

		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			parts := strings.Split(rule, "=")
			key := parts[0]
			param := ""
			if len(parts) > 1 {
				param = parts[1]
			}

			switch key {
			case "required":
				if isEmptyValue(value) {
					return fmt.Errorf("field '%s' is required", field.Name)
				}
			case "email":
				if value.Kind() == reflect.String && !emailRegex.MatchString(value.String()) {
					return fmt.Errorf("field '%s' must be a valid email", field.Name)
				}
			case "min":
				if err := checkMin(field.Name, value, param); err != nil {
					return err
				}
			case "max":
				if err := checkMax(field.Name, value, param); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkMin(fieldName string, v reflect.Value, param string) error {
	minVal, err := strconv.Atoi(param)
	if err != nil {
		return nil // Ignore invalid param
	}

	switch v.Kind() {
	case reflect.String:
		if len(v.String()) < minVal {
			return fmt.Errorf("field '%s' must be at least %d characters", fieldName, minVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() < int64(minVal) {
			return fmt.Errorf("field '%s' must be at least %d", fieldName, minVal)
		}
	}
	return nil
}

func checkMax(fieldName string, v reflect.Value, param string) error {
	maxVal, err := strconv.Atoi(param)
	if err != nil {
		return nil // Ignore invalid param
	}

	switch v.Kind() {
	case reflect.String:
		if len(v.String()) > maxVal {
			return fmt.Errorf("field '%s' must be at most %d characters", fieldName, maxVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > int64(maxVal) {
			return fmt.Errorf("field '%s' must be at most %d", fieldName, maxVal)
		}
	}
	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Map, reflect.Slice:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

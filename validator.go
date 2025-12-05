package ginji

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// ValidatorFunc is a custom validation function.
type ValidatorFunc func(value reflect.Value, param string) error

var (
	emailRegex    = regexp.MustCompile(`(?i)^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	numericRegex  = regexp.MustCompile(`^[0-9]+$`)
	alphanumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	// customValidators stores user-registered custom validators.
	customValidators = make(map[string]ValidatorFunc)
)

// RegisterValidator registers a custom validator function.
func RegisterValidator(tag string, fn ValidatorFunc) {
	customValidators[tag] = fn
}

// validateStruct checks struct tags for validation rules.
// Supported tags: required, email, url, alpha, numeric, alphanum, min, max, len, gt, gte, lt, lte, oneof, regex
func validateStruct(v any) error {
	return validateValue(reflect.ValueOf(v), "", make(map[uintptr]bool))
}

// validateValue validates a value recursively.
func validateValue(val reflect.Value, fieldPath string, visited map[uintptr]bool) error {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	// Prevent infinite loops from circular references
	if val.Kind() == reflect.Struct && val.CanAddr() {
		addr := val.Addr().Pointer()
		if visited[addr] {
			return nil
		}
		visited[addr] = true
		defer delete(visited, addr)
	}

	switch val.Kind() {
	case reflect.Struct:
		return validateStructFields(val, fieldPath, visited)
	case reflect.Slice, reflect.Array:
		return validateSliceOrArray(val, fieldPath, visited)
	case reflect.Map:
		return validateMap(val, fieldPath, visited)
	}

	return nil
}

// validateStructFields validates all fields in a struct.
func validateStructFields(val reflect.Value, parentPath string, visited map[uintptr]bool) error {
	t := val.Type()
	var validationErrors ValidationErrors

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)
		tag := field.Tag.Get("ginji")

		// Build field path
		fieldPath := field.Name
		if parentPath != "" {
			fieldPath = parentPath + "." + field.Name
		}

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		// Validate tags
		if tag != "" {
			if errs := validateFieldTags(fieldPath, value, tag); len(errs) > 0 {
				validationErrors = append(validationErrors, errs...)
			}
		}

		// Recursively validate nested structs, slices, arrays, maps
		if err := validateValue(value, fieldPath, visited); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				validationErrors = append(validationErrors, ve...)
			} else {
				return err
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

// validateSliceOrArray validates each element in a slice or array.
func validateSliceOrArray(val reflect.Value, fieldPath string, visited map[uintptr]bool) error {
	var validationErrors ValidationErrors

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		elemPath := fmt.Sprintf("%s[%d]", fieldPath, i)

		if err := validateValue(elem, elemPath, visited); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				validationErrors = append(validationErrors, ve...)
			} else {
				return err
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

// validateMap validates each value in a map.
func validateMap(val reflect.Value, fieldPath string, visited map[uintptr]bool) error {
	var validationErrors ValidationErrors

	for _, key := range val.MapKeys() {
		mapVal := val.MapIndex(key)
		elemPath := fmt.Sprintf("%s[%v]", fieldPath, key.Interface())

		if err := validateValue(mapVal, elemPath, visited); err != nil {
			if ve, ok := err.(ValidationErrors); ok {
				validationErrors = append(validationErrors, ve...)
			} else {
				return err
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}

// validateFieldTags validates a field based on its tags.
func validateFieldTags(fieldPath string, value reflect.Value, tag string) ValidationErrors {
	var errors ValidationErrors
	rules := strings.Split(tag, ",")

	for _, rule := range rules {
		parts := strings.SplitN(rule, "=", 2)
		key := strings.TrimSpace(parts[0])
		param := ""
		if len(parts) > 1 {
			param = strings.TrimSpace(parts[1])
		}

		// Check custom validators first
		if validator, ok := customValidators[key]; ok {
			if err := validator(value, param); err != nil {
				errors = append(errors, ValidationError{
					Field:   fieldPath,
					Message: err.Error(),
					Tag:     key,
					Value:   getValueInterface(value),
				})
			}
			continue
		}

		// Built-in validators
		if err := validateBuiltInRule(fieldPath, value, key, param); err != nil {
			errors = append(errors, *err)
		}
	}

	return errors
}

// validateBuiltInRule validates a single built-in rule.
func validateBuiltInRule(fieldPath string, value reflect.Value, key, param string) *ValidationError {
	switch key {
	case "required":
		if isEmptyValue(value) {
			return &ValidationError{
				Field:   fieldPath,
				Message: "field is required",
				Tag:     "required",
			}
		}

	case "email":
		if value.Kind() == reflect.String {
			if !emailRegex.MatchString(value.String()) {
				return &ValidationError{
					Field:   fieldPath,
					Message: "must be a valid email",
					Tag:     "email",
					Value:   value.String(),
				}
			}
		}

	case "url":
		if value.Kind() == reflect.String {
			if _, err := url.ParseRequestURI(value.String()); err != nil {
				return &ValidationError{
					Field:   fieldPath,
					Message: "must be a valid URL",
					Tag:     "url",
					Value:   value.String(),
				}
			}
		}

	case "alpha":
		if value.Kind() == reflect.String && value.String() != "" {
			if !alphaRegex.MatchString(value.String()) {
				return &ValidationError{
					Field:   fieldPath,
					Message: "must contain only letters",
					Tag:     "alpha",
					Value:   value.String(),
				}
			}
		}

	case "numeric":
		if value.Kind() == reflect.String && value.String() != "" {
			if !numericRegex.MatchString(value.String()) {
				return &ValidationError{
					Field:   fieldPath,
					Message: "must contain only numbers",
					Tag:     "numeric",
					Value:   value.String(),
				}
			}
		}

	case "alphanum":
		if value.Kind() == reflect.String && value.String() != "" {
			if !alphanumRegex.MatchString(value.String()) {
				return &ValidationError{
					Field:   fieldPath,
					Message: "must contain only letters and numbers",
					Tag:     "alphanum",
					Value:   value.String(),
				}
			}
		}

	case "min":
		if err := checkMin(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "min",
				Value:   getValueInterface(value),
			}
		}

	case "max":
		if err := checkMax(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "max",
				Value:   getValueInterface(value),
			}
		}

	case "len":
		if err := checkLen(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "len",
				Value:   getValueInterface(value),
			}
		}

	case "gt":
		if err := checkGt(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "gt",
				Value:   getValueInterface(value),
			}
		}

	case "gte":
		if err := checkGte(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "gte",
				Value:   getValueInterface(value),
			}
		}

	case "lt":
		if err := checkLt(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "lt",
				Value:   getValueInterface(value),
			}
		}

	case "lte":
		if err := checkLte(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "lte",
				Value:   getValueInterface(value),
			}
		}

	case "oneof":
		if err := checkOneOf(fieldPath, value, param); err != nil {
			return &ValidationError{
				Field:   fieldPath,
				Message: err.Error(),
				Tag:     "oneof",
				Value:   getValueInterface(value),
			}
		}

	case "regex":
		if value.Kind() == reflect.String {
			re, err := regexp.Compile(param)
			if err != nil {
				return &ValidationError{
					Field:   fieldPath,
					Message: "invalid regex pattern",
					Tag:     "regex",
				}
			}
			if !re.MatchString(value.String()) {
				return &ValidationError{
					Field:   fieldPath,
					Message: fmt.Sprintf("must match pattern: %s", param),
					Tag:     "regex",
					Value:   value.String(),
				}
			}
		}
	}

	return nil
}

func checkMin(fieldName string, v reflect.Value, param string) error {
	minVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil // Ignore invalid param
	}

	switch v.Kind() {
	case reflect.String:
		if float64(len(v.String())) < minVal {
			return fmt.Errorf("must be at least %.0f characters", minVal)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if float64(v.Len()) < minVal {
			return fmt.Errorf("must have at least %.0f items", minVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) < minVal {
			return fmt.Errorf("must be at least %.0f", minVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) < minVal {
			return fmt.Errorf("must be at least %.0f", minVal)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < minVal {
			return fmt.Errorf("must be at least %f", minVal)
		}
	}
	return nil
}

func checkMax(fieldName string, v reflect.Value, param string) error {
	maxVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil // Ignore invalid param
	}

	switch v.Kind() {
	case reflect.String:
		if float64(len(v.String())) > maxVal {
			return fmt.Errorf("must be at most %.0f characters", maxVal)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if float64(v.Len()) > maxVal {
			return fmt.Errorf("must have at most %.0f items", maxVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) > maxVal {
			return fmt.Errorf("must be at most %.0f", maxVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) > maxVal {
			return fmt.Errorf("must be at most %.0f", maxVal)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > maxVal {
			return fmt.Errorf("must be at most %f", maxVal)
		}
	}
	return nil
}

func checkLen(fieldName string, v reflect.Value, param string) error {
	lenVal, err := strconv.Atoi(param)
	if err != nil {
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		if len(v.String()) != lenVal {
			return fmt.Errorf("must be exactly %d characters", lenVal)
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() != lenVal {
			return fmt.Errorf("must have exactly %d items", lenVal)
		}
	}
	return nil
}

func checkGt(fieldName string, v reflect.Value, param string) error {
	gtVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) <= gtVal {
			return fmt.Errorf("must be greater than %.0f", gtVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) <= gtVal {
			return fmt.Errorf("must be greater than %.0f", gtVal)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() <= gtVal {
			return fmt.Errorf("must be greater than %f", gtVal)
		}
	}
	return nil
}

func checkGte(fieldName string, v reflect.Value, param string) error {
	gteVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) < gteVal {
			return fmt.Errorf("must be greater than or equal to %.0f", gteVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) < gteVal {
			return fmt.Errorf("must be greater than or equal to %.0f", gteVal)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() < gteVal {
			return fmt.Errorf("must be greater than or equal to %f", gteVal)
		}
	}
	return nil
}

func checkLt(fieldName string, v reflect.Value, param string) error {
	ltVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) >= ltVal {
			return fmt.Errorf("must be less than %.0f", ltVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) >= ltVal {
			return fmt.Errorf("must be less than %.0f", ltVal)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() >= ltVal {
			return fmt.Errorf("must be less than %f", ltVal)
		}
	}
	return nil
}

func checkLte(fieldName string, v reflect.Value, param string) error {
	lteVal, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return nil
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if float64(v.Int()) > lteVal {
			return fmt.Errorf("must be less than or equal to %.0f", lteVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if float64(v.Uint()) > lteVal {
			return fmt.Errorf("must be less than or equal to %.0f", lteVal)
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > lteVal {
			return fmt.Errorf("must be less than or equal to %f", lteVal)
		}
	}
	return nil
}

func checkOneOf(fieldName string, v reflect.Value, param string) error {
	if v.Kind() != reflect.String {
		return nil
	}

	options := strings.Split(param, " ")
	value := v.String()

	for _, option := range options {
		if value == option {
			return nil
		}
	}

	return fmt.Errorf("must be one of: %s", param)
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

func getValueInterface(v reflect.Value) any {
	if v.CanInterface() {
		return v.Interface()
	}
	return nil
}

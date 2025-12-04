package ginji

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// H is a shortcut for map[string]any
type H map[string]any

// jsonMarshal is a helper for marshaling JSON
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// jsonUnmarshal is a helper for unmarshaling JSON
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// bindMap binds a map of strings to a struct based on a tag.
func bindMap(data map[string][]string, v any, tagName string) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("bind target must be a struct")
	}

	t := val.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		if tag == "" {
			continue
		}

		// Check if the tag exists in the data
		if values, ok := data[tag]; ok && len(values) > 0 {
			fieldVal := val.Field(i)
			if fieldVal.CanSet() {
				// Use setField for proper type conversion
				if err := setField(fieldVal, values[0]); err != nil {
					return fmt.Errorf("failed to set field %s: %w", field.Name, err)
				}
			}
		}
	}

	return nil
}

// setField attempts to set the value of a reflect.Value field based on a string.
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int: %w", err)
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse uint: %w", err)
		}
		field.SetUint(u)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse bool: %w", err)
		}
		field.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float: %w", err)
		}
		field.SetFloat(f)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind().String())
	}
	return nil
}

// bindParams binds path parameters to a struct.
func bindParams(params map[string]string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("BindParams requires a non-nil pointer to a struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("BindParams requires a pointer to a struct")
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get the path tag
		paramName := field.Tag.Get("path")
		if paramName == "" {
			paramName = field.Tag.Get("param")
		}
		if paramName == "" {
			// Use field name as fallback (lowercase)
			paramName = strings.ToLower(field.Name)
		}

		// Get value from params
		value, exists := params[paramName]
		if !exists {
			continue
		}

		// Set the value
		if err := setField(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

// bindForm binds form data to a struct.
func bindForm(req *http.Request, v any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("BindForm requires a non-nil pointer to a struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("BindForm requires a pointer to a struct")
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		if !fieldValue.CanSet() {
			continue
		}

		// Get the form tag
		formName := field.Tag.Get("form")
		if formName == "" {
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				formName = parts[0]
			}
		}
		if formName == "" || formName == "-" {
			continue
		}

		// Get value from form
		value := req.Form.Get(formName)
		if value == "" {
			continue
		}

		// Set the value
		if err := setField(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

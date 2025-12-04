package ginji

import (
	"encoding/json"
	"fmt"
	"reflect"
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
				// Simple string setting for now
				if fieldVal.Kind() == reflect.String {
					fieldVal.SetString(values[0])
				}
				// TODO: Add support for other types (int, bool, etc.)
			}
		}
	}
	return nil
}

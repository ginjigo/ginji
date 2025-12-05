package ginji

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestValidateRequired(t *testing.T) {
	type User struct {
		Name string `ginji:"required"`
		Age  int    `ginji:"required"`
	}

	// Valid
	valid := User{Name: "John", Age: 30}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid - empty name
	invalid := User{Name: "", Age: 30}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for empty name")
	}
}

func TestValidateEmail(t *testing.T) {
	type Contact struct {
		Email string `ginji:"required,email"`
	}

	// Valid
	valid := Contact{Email: "test@example.com"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid
	invalid := Contact{Email: "not-an-email"}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for invalid email")
	}
}

func TestValidateURL(t *testing.T) {
	type Website struct {
		URL string `ginji:"url"`
	}

	// Valid
	valid := Website{URL: "https://example.com"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid
	invalid := Website{URL: "not a url"}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for invalid URL")
	}
}

func TestValidateAlpha(t *testing.T) {
	type Name struct {
		FirstName string `ginji:"alpha"`
	}

	valid := Name{FirstName: "John"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	invalid := Name{FirstName: "John123"}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for non-alpha string")
	}
}

func TestValidateNumeric(t *testing.T) {
	type Code struct {
		ZipCode string `ginji:"numeric"`
	}

	valid := Code{ZipCode: "12345"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	invalid := Code{ZipCode: "123abc"}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for non-numeric string")
	}
}

func TestValidateAlphanum(t *testing.T) {
	type Username struct {
		Name string `ginji:"alphanum"`
	}

	valid := Username{Name: "user123"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	invalid := Username{Name: "user-123"}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for non-alphanumeric string")
	}
}

func TestValidateMinMax(t *testing.T) {
	type Product struct {
		Name  string   `ginji:"min=3,max=50"`
		Price int      `ginji:"min=1,max=1000"`
		Tags  []string `ginji:"min=1,max=5"`
	}

	// Valid
	valid := Product{Name: "Laptop", Price: 500, Tags: []string{"electronics", "computers"}}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid - name too short
	invalid1 := Product{Name: "AB", Price: 500, Tags: []string{"tag"}}
	if err := validateStruct(&invalid1); err == nil {
		t.Error("Expected validation error for name too short")
	}

	// Invalid - price too low
	invalid2 := Product{Name: "Test", Price: 0, Tags: []string{"tag"}}
	if err := validateStruct(&invalid2); err == nil {
		t.Error("Expected validation error for price too low")
	}

	// Invalid - too many tags
	invalid3 := Product{Name: "Test", Price: 100, Tags: []string{"1", "2", "3", "4", "5", "6"}}
	if err := validateStruct(&invalid3); err == nil {
		t.Error("Expected validation error for too many tags")
	}
}

func TestValidateLen(t *testing.T) {
	type Code struct {
		PostalCode string `ginji:"len=5"`
	}

	valid := Code{PostalCode: "12345"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	invalid := Code{PostalCode: "123"}
	if err := validateStruct(&invalid); err == nil {
		t.Error("Expected validation error for incorrect length")
	}
}

func TestValidateGtGteLtLte(t *testing.T) {
	type Numbers struct {
		Greater      int `ginji:"gt=10"`
		GreaterEqual int `ginji:"gte=10"`
		Lesser       int `ginji:"lt=100"`
		LesserEqual  int `ginji:"lte=100"`
	}

	// Valid
	valid := Numbers{Greater: 11, GreaterEqual: 10, Lesser: 99, LesserEqual: 100}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid - greater
	invalid1 := Numbers{Greater: 10, GreaterEqual: 10, Lesser: 99, LesserEqual: 100}
	if err := validateStruct(&invalid1); err == nil {
		t.Error("Expected validation error for gt")
	}

	// Invalid - lesser
	invalid2 := Numbers{Greater: 11, GreaterEqual: 10, Lesser: 100, LesserEqual: 100}
	if err := validateStruct(&invalid2); err == nil {
		t.Error("Expected validation error for lt")
	}
}

func TestValidateOneOf(t *testing.T) {
	type Status struct {
		State string `ginji:"oneof=active inactive pending"`
	}

	// Valid
	valid := Status{State: "active"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid
	invalid := Status{State: "deleted"}
	if err := validateStruct(&invalid); err == nil {
		t.Error("Expected validation error for invalid oneof value")
	}
}

func TestValidateRegex(t *testing.T) {
	type Phone struct {
		Number string `ginji:"regex=^\\d{3}-\\d{3}-\\d{4}$"`
	}

	// Valid
	valid := Phone{Number: "123-456-7890"}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid
	invalid := Phone{Number: "1234567890"}
	if err := validateStruct(&invalid); err == nil {
		t.Error("Expected validation error for regex mismatch")
	}
}

func TestValidateNestedStruct(t *testing.T) {
	type Address struct {
		Street string `ginji:"required,min=5"`
		City   string `ginji:"required"`
		Zip    string `ginji:"required,numeric,len=5"`
	}

	type User struct {
		Name    string  `ginji:"required"`
		Address Address `ginji:"required"`
	}

	// Valid
	valid := User{
		Name: "John Doe",
		Address: Address{
			Street: "123 Main St",
			City:   "New York",
			Zip:    "10001",
		},
	}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid - nested field validation
	invalid := User{
		Name: "John Doe",
		Address: Address{
			Street: "123", // Too short
			City:   "New York",
			Zip:    "10001",
		},
	}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for nested field")
	}

	// Check if error contains nested field path
	if verrs, ok := err.(ValidationErrors); ok {
		if len(verrs) == 0 || !contains(verrs[0].Field, "Address") {
			t.Error("Expected error to include nested field path")
		}
	}
}

func TestValidateSlice(t *testing.T) {
	type Item struct {
		Name string `ginji:"required,min=2"`
	}

	type Cart struct {
		Items []Item `ginji:"min=1"`
	}

	// Valid
	valid := Cart{
		Items: []Item{
			{Name: "Item1"},
			{Name: "Item2"},
		},
	}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid - item in slice
	invalid := Cart{
		Items: []Item{
			{Name: "Item1"},
			{Name: "I"}, // Too short
		},
	}
	err := validateStruct(&invalid)
	if err == nil {
		t.Error("Expected validation error for item in slice")
	}

	// Check if error contains array index
	if verrs, ok := err.(ValidationErrors); ok {
		if len(verrs) == 0 || !contains(verrs[0].Field, "[1]") {
			t.Error("Expected error to include array index")
		}
	}
}

func TestCustomValidator(t *testing.T) {
	// Register custom validator
	RegisterValidator("even", func(value reflect.Value, param string) error {
		if value.Kind() == reflect.Int {
			if value.Int()%2 != 0 {
				return fmt.Errorf("must be an even number")
			}
		}
		return nil
	})

	type Numbers struct {
		EvenNumber int `ginji:"even"`
	}

	// Valid
	valid := Numbers{EvenNumber: 10}
	if err := validateStruct(&valid); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Invalid
	invalid := Numbers{EvenNumber: 11}
	if err := validateStruct(&invalid); err == nil {
		t.Error("Expected validation error for odd number")
	}

	// Cleanup
	delete(customValidators, "even")
}

func TestValidationErrorsCollection(t *testing.T) {
	type Form struct {
		Email string `ginji:"required,email"`
		Age   int    `ginji:"required,min=18"`
		Name  string `ginji:"required,min=2"`
	}

	invalid := Form{
		Email: "invalid",
		Age:   15,
		Name:  "",
	}

	err := validateStruct(&invalid)
	if err == nil {
		t.Fatal("Expected validation errors")
	}

	verrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatal("Expected ValidationErrors type")
	}

	// Should have multiple errors
	if len(verrs) < 3 {
		t.Errorf("Expected at least 3 validation errors, got %d", len(verrs))
	}
}

func TestValidateMap(t *testing.T) {
	type Config struct {
		Settings map[string]string
	}

	// Maps themselves don't have validation tags, but their values can be validated
	// if they are structs
	config := Config{
		Settings: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	if err := validateStruct(&config); err != nil {
		t.Errorf("Expected no error for valid map, got: %v", err)
	}
}

func TestCircularReferenceProtection(t *testing.T) {
	type Node struct {
		Value string `ginji:"required"`
		Next  *Node
	}

	// Create circular reference
	node1 := &Node{Value: "first"}
	node2 := &Node{Value: "second"}
	node1.Next = node2
	node2.Next = node1

	// Should not cause infinite loop
	if err := validateStruct(node1); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			strings.Contains(s, substr)))
}

package ginji

import (
	"strings"
	"testing"
)

// TestValidationNestedStructs tests validation of deeply nested structures
func TestValidationNestedStructs(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type Person struct {
		Name    string  `validate:"required"`
		Address Address `validate:"required"`
	}

	tests := []struct {
		name    string
		person  Person
		wantErr bool
	}{
		{
			name: "Valid nested struct",
			person: Person{
				Name:    "John",
				Address: Address{Street: "123 Main", City: "NYC"},
			},
			wantErr: false,
		},
		{
			name: "Missing nested field",
			person: Person{
				Name:    "John",
				Address: Address{Street: "123 Main"}, // Missing City
			},
			wantErr: true,
		},
		{
			name: "Missing top-level field",
			person: Person{
				Address: Address{Street: "123 Main", City: "NYC"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(&tt.person)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidationEmailEdgeCases tests email validation edge cases
func TestValidationEmailEdgeCases(t *testing.T) {
	type User struct {
		Email string `validate:"email"`
	}

	tests := []struct {
		email   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"user+tag@example.com", false},
		{"user.name@example.com", false},
		{"user@subdomain.example.com", false},
		{"", true},                  // Empty
		{"notanemail", true},        //No @
		{"@example.com", true},      // Missing local part
		{"user@", true},             // Missing domain
		{"user @example.com", true}, // Space in address
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			user := User{Email: tt.email}
			err := validateStruct(&user)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() for email %q, error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

// TestValidationURLEdgeCases tests URL validation edge cases
func TestValidationURLEdgeCases(t *testing.T) {
	type Resource struct {
		URL string `validate:"url"`
	}

	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com", false},
		{"http://example.com", false},
		{"https://example.com/path", false},
		{"https://example.com:8080/path", false},
		{"", true},                  // Empty
		{"notaurl", true},           // No scheme
		{"htp://example.com", true}, // Invalid scheme
		{"https://", true},          // Missing host
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			resource := Resource{URL: tt.url}
			err := validateStruct(&resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() for URL %q, error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

// TestValidationNumberRanges tests min/max/gt/gte/lt/lte validations
func TestValidationNumberRanges(t *testing.T) {
	type Product struct {
		Price    int `validate:"min=0,max=1000"`
		Quantity int `validate:"gt=0"`
		Discount int `validate:"gte=0,lte=100"`
	}

	tests := []struct {
		name    string
		product Product
		wantErr bool
	}{
		{
			name:    "Valid product",
			product: Product{Price: 500, Quantity: 10, Discount: 50},
			wantErr: false,
		},
		{
			name:    "Price too low",
			product: Product{Price: -1, Quantity: 10, Discount: 50},
			wantErr: true,
		},
		{
			name:    "Price too high",
			product: Product{Price: 1001, Quantity: 10, Discount: 50},
			wantErr: true,
		},
		{
			name:    "Quantity zero (not allowed by gt)",
			product: Product{Price: 500, Quantity: 0, Discount: 50},
			wantErr: true,
		},
		{
			name:    "Discount negative",
			product: Product{Price: 500, Quantity: 10, Discount: -1},
			wantErr: true,
		},
		{
			name:    "Discount boundary (100 is valid for lte)",
			product: Product{Price: 500, Quantity: 10, Discount: 100},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(&tt.product)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidationStringLength tests len, min, max for strings
func TestValidationStringLength(t *testing.T) {
	type Post struct {
		Title       string `validate:"len=10"`
		Description string `validate:"min=5,max=100"`
	}

	tests := []struct {
		name    string
		post    Post
		wantErr bool
	}{
		{
			name:    "Valid lengths",
			post:    Post{Title: "1234567890", Description: "Hello there"},
			wantErr: false,
		},
		{
			name:    "Title too short",
			post:    Post{Title: "short", Description: "Hello there"},
			wantErr: true,
		},
		{
			name:    "Title too long",
			post:    Post{Title: "12345678901", Description: "Hello there"},
			wantErr: true,
		},
		{
			name:    "Description too short",
			post:    Post{Title: "1234567890", Description: "Hi"},
			wantErr: true,
		},
		{
			name:    "Description too long",
			post:    Post{Title: "1234567890", Description: strings.Repeat("a", 101)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(&tt.post)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidationOneOf tests oneof validation
func TestValidationOneOf(t *testing.T) {
	type Config struct {
		Environment string `validate:"oneof=development staging production"`
	}

	tests := []struct {
		env     string
		wantErr bool
	}{
		{"development", false},
		{"staging", false},
		{"production", false},
		{"test", true},
		{"", true},
		{"Development", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.env, func(t *testing.T) {
			config := Config{Environment: tt.env}
			err := validateStruct(&config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() for env %q, error = %v, wantErr %v", tt.env, err, tt.wantErr)
			}
		})
	}
}

// TestValidationRegex tests regex pattern validation
func TestValidationRegex(t *testing.T) {
	type Code struct {
		ZipCode string `validate:"regex=^[0-9]{5}$"`
	}

	tests := []struct {
		zipCode string
		wantErr bool
	}{
		{"12345", false},
		{"00000", false},
		{"1234", true},   // Too short
		{"123456", true}, // Too long
		{"abcde", true},  // Not digits
		{"", true},       // Empty
	}

	for _, tt := range tests {
		t.Run(tt.zipCode, func(t *testing.T) {
			code := Code{ZipCode: tt.zipCode}
			err := validateStruct(&code)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() for zipCode %q, error = %v, wantErr %v", tt.zipCode, err, tt.wantErr)
			}
		})
	}
}

// TestValidationSliceAndMap tests validation of slices and maps
func TestValidationSliceAndMap(t *testing.T) {
	type Data struct {
		Tags  []string          `validate:"required"`
		Props map[string]string `validate:"required"`
	}

	tests := []struct {
		name    string
		data    Data
		wantErr bool
	}{
		{
			name:    "Valid with data",
			data:    Data{Tags: []string{"a"}, Props: map[string]string{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "Empty slice",
			data:    Data{Tags: []string{}, Props: map[string]string{"key": "value"}},
			wantErr: true,
		},
		{
			name:    "Empty map",
			data:    Data{Tags: []string{"a"}, Props: map[string]string{}},
			wantErr: true,
		},
		{
			name:    "Nil slice",
			data:    Data{Tags: nil, Props: map[string]string{"key": "value"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(&tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidationPointerFields tests validation of pointer fields
func TestValidationPointerFields(t *testing.T) {
	type Optional struct {
		Age *int `validate:"gt=0"`
	}

	validAge := 25
	invalidAge := -5

	tests := []struct {
		name    string
		opt     Optional
		wantErr bool
	}{
		{
			name:    "Valid pointer value",
			opt:     Optional{Age: &validAge},
			wantErr: false,
		},
		{
			name:    "Invalid pointer value",
			opt:     Optional{Age: &invalidAge},
			wantErr: true,
		},
		{
			name:    "Nil pointer (should be valid since not required)",
			opt:     Optional{Age: nil},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStruct(&tt.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

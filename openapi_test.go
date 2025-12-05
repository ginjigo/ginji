package ginji

import (
	"reflect"
	"testing"
)

func TestOpenAPIGeneration(t *testing.T) {
	app := New()

	type TestRequest struct {
		Name string `json:"name" validate:"required" description:"Test name" example:"test"`
	}

	type TestResponse struct {
		ID   int    `json:"id" description:"Test ID"`
		Name string `json:"name" description:"Test name"`
	}

	// Define routes with metadata
	app.Get("/test", func(c *Context) error {
		return c.JSON(200, TestResponse{ID: 1, Name: "test"})
	}).
		Summary("Test endpoint").
		Description("A test endpoint for OpenAPI generation").
		Tags("test").
		Response(200, TestResponse{})

	app.Post("/test", func(c *Context) error {
		return c.JSON(201, TestResponse{ID: 1, Name: "test"})
	}).
		Summary("Create test").
		Tags("test").
		Request(TestRequest{}).
		Response(201, TestResponse{})

	// Generate OpenAPI spec
	spec := app.GenerateOpenAPI(OpenAPIConfig{
		Title:       "Test API",
		Description: "Test API Description",
		Version:     "1.0.0",
	})

	// Verify spec basics
	if spec.OpenAPI != "3.1.0" {
		t.Errorf("Expected OpenAPI version 3.1.0, got %s", spec.OpenAPI)
	}

	if spec.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got %s", spec.Info.Title)
	}

	if spec.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", spec.Info.Version)
	}

	// Verify paths
	if len(spec.Paths) == 0 {
		t.Fatal("Expected paths to be generated")
	}

	testPath, exists := spec.Paths["/test"]
	if !exists {
		t.Fatal("Expected /test path to exist")
	}

	// Verify GET operation
	if testPath.Get == nil {
		t.Fatal("Expected GET operation to exist")
	}

	if testPath.Get.Summary != "Test endpoint" {
		t.Errorf("Expected summary 'Test endpoint', got %s", testPath.Get.Summary)
	}

	if len(testPath.Get.Tags) == 0 || testPath.Get.Tags[0] != "test" {
		t.Error("Expected 'test' tag")
	}

	// Verify POST operation
	if testPath.Post == nil {
		t.Fatal("Expected POST operation to exist")
	}

	if testPath.Post.RequestBody == nil {
		t.Fatal("Expected request body for POST")
	}

	// Verify components/schemas
	if spec.Components == nil || spec.Components.Schemas == nil {
		t.Fatal("Expected components schemas to exist")
	}

	if _, exists := spec.Components.Schemas["TestRequest"]; !exists {
		t.Error("Expected TestRequest schema")
	}

	if _, exists := spec.Components.Schemas["TestResponse"]; !exists {
		t.Error("Expected TestResponse schema")
	}
}

func TestSchemaGeneration(t *testing.T) {
	type TestStruct struct {
		Name     string  `json:"name" validate:"required" description:"Name field" example:"John"`
		Age      int     `json:"age" description:"Age field" example:"30"`
		Email    string  `json:"email" validate:"email" description:"Email field"`
		Optional *string `json:"optional,omitempty" description:"Optional field"`
	}

	schemas := make(map[string]*OpenAPISchema)
	schema := generateSchema(reflect.TypeOf(TestStruct{}), schemas)

	// For named types, it returns a reference, so get the actual schema from the map
	if schema.Ref != "" {
		// Schema was stored in components
		actualSchema, exists := schemas["TestStruct"]
		if !exists {
			t.Fatal("Expected TestStruct schema to be in components")
		}
		schema = actualSchema
	}

	// Verify schema type
	if schema.Type != "object" {
		t.Errorf("Expected type 'object', got %s", schema.Type)
	}

	// Verify properties
	if schema.Properties == nil {
		t.Fatal("Expected properties to exist")
	}

	// Check name field
	nameSchema, exists := schema.Properties["name"]
	if !exists {
		t.Fatal("Expected 'name' property")
	}

	if nameSchema.Type != "string" {
		t.Errorf("Expected name type 'string', got %s", nameSchema.Type)
	}

	if nameSchema.Description != "Name field" {
		t.Errorf("Expected description 'Name field', got %s", nameSchema.Description)
	}

	if nameSchema.Example != "John" {
		t.Errorf("Expected example 'John', got %v", nameSchema.Example)
	}

	// Check age field
	ageSchema, exists := schema.Properties["age"]
	if !exists {
		t.Fatal("Expected 'age' property")
	}

	if ageSchema.Type != "integer" {
		t.Errorf("Expected age type 'integer', got %s", ageSchema.Type)
	}

	// Verify required fields (only 'name' has validate:"required")
	if len(schema.Required) != 1 {
		t.Errorf("Expected 1 required field, got %d", len(schema.Required))
	}

	if len(schema.Required) > 0 && schema.Required[0] != "name" {
		t.Errorf("Expected 'name' to be required, got %s", schema.Required[0])
	}
}

func TestRouteMetadata(t *testing.T) {
	app := New()

	type Req struct {
		Name string `json:"name"`
	}

	type Resp struct {
		ID int `json:"id"`
	}

	_ = app.Post("/test", func(c *Context) error { return nil }).
		Summary("Test summary").
		Description("Test description").
		Tags("tag1", "tag2").
		OperationID("testOp").
		Request(Req{}).
		Response(200, Resp{}).
		Response(400, map[string]string{}).
		Deprecated()

	// Verify metadata was set
	key := "POST-/test"
	meta := app.router.getRouteMetadata(key)

	if meta.Summary != "Test summary" {
		t.Errorf("Expected summary 'Test summary', got %s", meta.Summary)
	}

	if meta.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", meta.Description)
	}

	if len(meta.Tags) != 2 || meta.Tags[0] != "tag1" || meta.Tags[1] != "tag2" {
		t.Errorf("Expected tags [tag1, tag2], got %v", meta.Tags)
	}

	if meta.OperationID != "testOp" {
		t.Errorf("Expected operation ID 'testOp', got %s", meta.OperationID)
	}

	if !meta.Deprecated {
		t.Error("Expected route to be deprecated")
	}

	if meta.RequestType == nil {
		t.Error("Expected request type to be set")
	}

	if len(meta.Responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(meta.Responses))
	}
}

func TestExtractPathParameters(t *testing.T) {
	tests := []struct {
		pattern  string
		expected []string
	}{
		{"/users/:id", []string{"id"}},
		{"/users/:id/posts/:postId", []string{"id", "postId"}},
		{"/static/*filepath", []string{}},
		{"/api/v1/users/:userId/comments/:commentId", []string{"userId", "commentId"}},
		{"/users", []string{}},
	}

	for _, tt := range tests {
		result := extractPathParameters(tt.pattern)
		if len(result) != len(tt.expected) {
			t.Errorf("For pattern %s: expected %d params, got %d", tt.pattern, len(tt.expected), len(result))
			continue
		}

		for i, param := range result {
			if param != tt.expected[i] {
				t.Errorf("For pattern %s: expected param %s, got %s", tt.pattern, tt.expected[i], param)
			}
		}
	}
}

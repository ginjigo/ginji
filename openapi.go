package ginji

import (
	"encoding/json"
	"reflect"
	"strings"
)

// OpenAPIInfo represents API information.
type OpenAPIInfo struct {
	Title          string                 `json:"title"`
	Description    string                 `json:"description,omitempty"`
	Version        string                 `json:"version"`
	TermsOfService string                 `json:"termsOfService,omitempty"`
	Contact        *OpenAPIContact        `json:"contact,omitempty"`
	License        *OpenAPILicense        `json:"license,omitempty"`
	Extensions     map[string]interface{} `json:"-"`
}

// OpenAPIContact represents contact information.
type OpenAPIContact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// OpenAPILicense represents license information.
type OpenAPILicense struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// OpenAPIServer represents a server.
type OpenAPIServer struct {
	URL         string                           `json:"url"`
	Description string                           `json:"description,omitempty"`
	Variables   map[string]OpenAPIServerVariable `json:"variables,omitempty"`
}

// OpenAPIServerVariable represents a server variable.
type OpenAPIServerVariable struct {
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// OpenAPISpec represents the OpenAPI specification.
type OpenAPISpec struct {
	OpenAPI      string                     `json:"openapi"`
	Info         OpenAPIInfo                `json:"info"`
	Servers      []OpenAPIServer            `json:"servers,omitempty"`
	Paths        map[string]OpenAPIPathItem `json:"paths"`
	Components   *OpenAPIComponents         `json:"components,omitempty"`
	Security     []map[string][]string      `json:"security,omitempty"`
	Tags         []OpenAPITag               `json:"tags,omitempty"`
	ExternalDocs *OpenAPIExternalDocs       `json:"externalDocs,omitempty"`
}

// OpenAPIPathItem represents a path item.
type OpenAPIPathItem struct {
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Get         *OpenAPIOperation  `json:"get,omitempty"`
	Post        *OpenAPIOperation  `json:"post,omitempty"`
	Put         *OpenAPIOperation  `json:"put,omitempty"`
	Delete      *OpenAPIOperation  `json:"delete,omitempty"`
	Patch       *OpenAPIOperation  `json:"patch,omitempty"`
	Options     *OpenAPIOperation  `json:"options,omitempty"`
	Head        *OpenAPIOperation  `json:"head,omitempty"`
	Parameters  []OpenAPIParameter `json:"parameters,omitempty"`
}

// OpenAPIOperation represents an operation.
type OpenAPIOperation struct {
	Tags        []string                   `json:"tags,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
	Security    []map[string][]string      `json:"security,omitempty"`
	Deprecated  bool                       `json:"deprecated,omitempty"`
}

// OpenAPIParameter represents a parameter.
type OpenAPIParameter struct {
	Name        string                 `json:"name"`
	In          string                 `json:"in"` // query, header, path, cookie
	Description string                 `json:"description,omitempty"`
	Required    bool                   `json:"required,omitempty"`
	Deprecated  bool                   `json:"deprecated,omitempty"`
	Schema      *OpenAPISchema         `json:"schema,omitempty"`
	Example     interface{}            `json:"example,omitempty"`
	Examples    map[string]interface{} `json:"examples,omitempty"`
}

// OpenAPIRequestBody represents a request body.
type OpenAPIRequestBody struct {
	Description string                      `json:"description,omitempty"`
	Content     map[string]OpenAPIMediaType `json:"content"`
	Required    bool                        `json:"required,omitempty"`
}

// OpenAPIResponse represents a response.
type OpenAPIResponse struct {
	Description string                      `json:"description"`
	Content     map[string]OpenAPIMediaType `json:"content,omitempty"`
	Headers     map[string]OpenAPIHeader    `json:"headers,omitempty"`
}

// OpenAPIMediaType represents a media type.
type OpenAPIMediaType struct {
	Schema   *OpenAPISchema         `json:"schema,omitempty"`
	Example  interface{}            `json:"example,omitempty"`
	Examples map[string]interface{} `json:"examples,omitempty"`
}

// OpenAPIHeader represents a header.
type OpenAPIHeader struct {
	Description string         `json:"description,omitempty"`
	Required    bool           `json:"required,omitempty"`
	Schema      *OpenAPISchema `json:"schema,omitempty"`
}

// OpenAPISchema represents a JSON Schema.
type OpenAPISchema struct {
	Type                 string                    `json:"type,omitempty"`
	Format               string                    `json:"format,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Enum                 []interface{}             `json:"enum,omitempty"`
	Default              interface{}               `json:"default,omitempty"`
	Example              interface{}               `json:"example,omitempty"`
	Properties           map[string]*OpenAPISchema `json:"properties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Items                *OpenAPISchema            `json:"items,omitempty"`
	AdditionalProperties interface{}               `json:"additionalProperties,omitempty"`
	Nullable             bool                      `json:"nullable,omitempty"`
	ReadOnly             bool                      `json:"readOnly,omitempty"`
	WriteOnly            bool                      `json:"writeOnly,omitempty"`
	Deprecated           bool                      `json:"deprecated,omitempty"`
	Ref                  string                    `json:"$ref,omitempty"`
	MinLength            *int                      `json:"minLength,omitempty"`
	MaxLength            *int                      `json:"maxLength,omitempty"`
	Minimum              *float64                  `json:"minimum,omitempty"`
	Maximum              *float64                  `json:"maximum,omitempty"`
	Pattern              string                    `json:"pattern,omitempty"`
}

// OpenAPIComponents represents reusable components.
type OpenAPIComponents struct {
	Schemas         map[string]*OpenAPISchema        `json:"schemas,omitempty"`
	Responses       map[string]OpenAPIResponse       `json:"responses,omitempty"`
	Parameters      map[string]OpenAPIParameter      `json:"parameters,omitempty"`
	RequestBodies   map[string]OpenAPIRequestBody    `json:"requestBodies,omitempty"`
	Headers         map[string]OpenAPIHeader         `json:"headers,omitempty"`
	SecuritySchemes map[string]OpenAPISecurityScheme `json:"securitySchemes,omitempty"`
}

// OpenAPISecurityScheme represents a security scheme.
type OpenAPISecurityScheme struct {
	Type             string `json:"type"` // apiKey, http, oauth2, openIdConnect
	Description      string `json:"description,omitempty"`
	Name             string `json:"name,omitempty"`             // For apiKey
	In               string `json:"in,omitempty"`               // For apiKey: query, header, cookie
	Scheme           string `json:"scheme,omitempty"`           // For http: basic, bearer
	BearerFormat     string `json:"bearerFormat,omitempty"`     // For http bearer
	OpenIdConnectURL string `json:"openIdConnectUrl,omitempty"` // For openIdConnect
}

// OpenAPITag represents a tag.
type OpenAPITag struct {
	Name         string               `json:"name"`
	Description  string               `json:"description,omitempty"`
	ExternalDocs *OpenAPIExternalDocs `json:"externalDocs,omitempty"`
}

// OpenAPIExternalDocs represents external documentation.
type OpenAPIExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// OpenAPIConfig represents OpenAPI configuration.
type OpenAPIConfig struct {
	Title           string
	Description     string
	Version         string
	TermsOfService  string
	Contact         *OpenAPIContact
	License         *OpenAPILicense
	Servers         []OpenAPIServer
	SecuritySchemes map[string]OpenAPISecurityScheme
	Tags            []OpenAPITag
}

// GenerateOpenAPI generates an OpenAPI specification from the router.
func (engine *Engine) GenerateOpenAPI(config OpenAPIConfig) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: OpenAPIInfo{
			Title:          config.Title,
			Description:    config.Description,
			Version:        config.Version,
			TermsOfService: config.TermsOfService,
			Contact:        config.Contact,
			License:        config.License,
		},
		Servers: config.Servers,
		Paths:   make(map[string]OpenAPIPathItem),
		Components: &OpenAPIComponents{
			Schemas:         make(map[string]*OpenAPISchema),
			SecuritySchemes: config.SecuritySchemes,
		},
		Tags: config.Tags,
	}

	// Scan all routes and generate paths
	engine.router.generatePaths(spec)

	return spec
}

// generatePaths generates OpenAPI paths from router.
func (r *Router) generatePaths(spec *OpenAPISpec) {
	for method, root := range r.roots {
		r.traverseNode(root, method, "", spec)
	}
}

// traverseNode traverses the trie and generates path items.
func (r *Router) traverseNode(node *node, method string, currentPath string, spec *OpenAPISpec) {
	if node == nil {
		return
	}

	// Build current path
	if node.part != "" {
		if currentPath == "" {
			currentPath = "/" + node.part
		} else {
			currentPath = currentPath + "/" + node.part
		}
	}

	// If this node has a pattern (is a route endpoint)
	if node.pattern != "" {
		// Get or create path item
		pathItem, exists := spec.Paths[node.pattern]
		if !exists {
			pathItem = OpenAPIPathItem{}
		}

		// Get route metadata
		key := method + "-" + node.pattern
		metadata := r.getRouteMetadata(key)

		// Create operation
		operation := &OpenAPIOperation{
			Summary:     metadata.Summary,
			Description: metadata.Description,
			Tags:        metadata.Tags,
			OperationID: metadata.OperationID,
			Responses:   make(map[string]OpenAPIResponse),
			Deprecated:  metadata.Deprecated,
		}

		// Add path parameters
		params := extractPathParameters(node.pattern)
		for _, param := range params {
			operation.Parameters = append(operation.Parameters, OpenAPIParameter{
				Name:     param,
				In:       "path",
				Required: true,
				Schema: &OpenAPISchema{
					Type: "string",
				},
			})
		}

		// Add request body if specified
		if metadata.RequestType != nil {
			schema := generateSchema(metadata.RequestType, spec.Components.Schemas)
			operation.RequestBody = &OpenAPIRequestBody{
				Required: true,
				Content: map[string]OpenAPIMediaType{
					"application/json": {
						Schema: schema,
					},
				},
			}
		}

		// Add responses
		if len(metadata.Responses) > 0 {
			for code, respType := range metadata.Responses {
				schema := generateSchema(respType, spec.Components.Schemas)
				operation.Responses[code] = OpenAPIResponse{
					Description: getResponseDescription(code),
					Content: map[string]OpenAPIMediaType{
						"application/json": {
							Schema: schema,
						},
					},
				}
			}
		} else {
			// Default response
			operation.Responses["200"] = OpenAPIResponse{
				Description: "Successful response",
			}
		}

		// Set operation on path item
		switch strings.ToUpper(method) {
		case "GET":
			pathItem.Get = operation
		case "POST":
			pathItem.Post = operation
		case "PUT":
			pathItem.Put = operation
		case "DELETE":
			pathItem.Delete = operation
		case "PATCH":
			pathItem.Patch = operation
		}

		spec.Paths[node.pattern] = pathItem
	}

	// Traverse children
	for _, child := range node.children {
		r.traverseNode(child, method, currentPath, spec)
	}
}

// extractPathParameters extracts parameter names from a path pattern.
func extractPathParameters(pattern string) []string {
	var params []string
	parts := strings.Split(pattern, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, part[1:])
		}
	}
	return params
}

// generateSchema generates an OpenAPI schema from a Go type.
func generateSchema(t reflect.Type, schemas map[string]*OpenAPISchema) *OpenAPISchema {
	if t == nil {
		return nil
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if already in components
	typeName := t.Name()
	if typeName != "" {
		if _, exists := schemas[typeName]; exists {
			return &OpenAPISchema{
				Ref: "#/components/schemas/" + typeName,
			}
		}
	}

	schema := &OpenAPISchema{}

	switch t.Kind() {
	case reflect.String:
		schema.Type = "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = "integer"
	case reflect.Float32, reflect.Float64:
		schema.Type = "number"
	case reflect.Bool:
		schema.Type = "boolean"
	case reflect.Slice, reflect.Array:
		schema.Type = "array"
		schema.Items = generateSchema(t.Elem(), schemas)
	case reflect.Map:
		schema.Type = "object"
		schema.AdditionalProperties = generateSchema(t.Elem(), schemas)
	case reflect.Struct:
		schema.Type = "object"
		schema.Properties = make(map[string]*OpenAPISchema)
		var required []string

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			fieldName := field.Name
			if jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" {
					fieldName = parts[0]
				}
			}

			fieldSchema := generateSchema(field.Type, schemas)

			// Add description from tag
			if desc := field.Tag.Get("description"); desc != "" {
				fieldSchema.Description = desc
			}

			// Add example from tag
			if example := field.Tag.Get("example"); example != "" {
				fieldSchema.Example = example
			}

			// Check if field is required
			if validateTag := field.Tag.Get("validate"); validateTag != "" {
				if strings.Contains(validateTag, "required") {
					required = append(required, fieldName)
				}
			}

			schema.Properties[fieldName] = fieldSchema
		}

		if len(required) > 0 {
			schema.Required = required
		}

		// Add to components if it has a name
		if typeName != "" {
			schemas[typeName] = schema
			return &OpenAPISchema{
				Ref: "#/components/schemas/" + typeName,
			}
		}
	}

	return schema
}

// getResponseDescription returns a default description for a status code.
func getResponseDescription(code string) string {
	descriptions := map[string]string{
		"200": "Successful response",
		"201": "Created",
		"204": "No content",
		"400": "Bad request",
		"401": "Unauthorized",
		"403": "Forbidden",
		"404": "Not found",
		"500": "Internal server error",
	}
	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return "Response"
}

// MarshalJSON customizes JSON marshaling for OpenAPISpec.
func (spec *OpenAPISpec) MarshalJSON() ([]byte, error) {
	type Alias OpenAPISpec
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(spec),
	})
}

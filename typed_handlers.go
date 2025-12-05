package ginji

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// TypedHandler is a generic handler with typed request and response.
// It provides compile-time type safety for request and response handling.
type TypedHandler[Req any, Res any] func(*Context, Req) (Res, error)

// EmptyRequest is used when a handler doesn't need a request body.
type EmptyRequest struct{}

// TypedHandlerFunc wraps a typed handler for use with standard routing.
// It automatically handles request binding, validation, and response marshaling.
func TypedHandlerFunc[Req any, Res any](handler TypedHandler[Req, Res]) Handler {
	// Cache type checks at handler creation time
	var emptyReq Req
	var emptyRes Res

	reqType := reflect.TypeOf(emptyReq)
	resType := reflect.TypeOf(emptyRes)
	emptyReqType := reflect.TypeOf(EmptyRequest{})

	isEmptyReq := reqType == emptyReqType
	isEmptyRes := resType == emptyReqType

	reqTypeName := getTypeName(reqType)

	return func(c *Context) error {
		var req Req

		// Skip binding for EmptyRequest
		if !isEmptyReq {
			// Attempt to bind the request with detailed error messages
			if err := bindTypedRequest(c, &req); err != nil {
				errorMsg := fmt.Sprintf("Failed to bind request to type %s: %v", reqTypeName, err)
				c.AbortWithError(StatusBadRequest, NewHTTPError(StatusBadRequest, errorMsg))
				return nil
			}

			// Validate the bound request
			if err := validateStruct(req); err != nil {
				c.AbortWithError(StatusUnprocessableEntity, err)
				return nil
			}
		}

		// Call the typed handler
		res, err := handler(c, req)
		if err != nil {
			// Check if it's already an HTTPError
			if httpErr, ok := err.(*HTTPError); ok {
				c.AbortWithError(httpErr.Code, httpErr)
			} else {
				c.AbortWithError(StatusInternalServerError, err)
			}
			return nil
		}

		// Skip response for EmptyRequest
		if isEmptyRes {
			// No response body, just return 204 No Content or keep existing status
			if c.StatusCode() == StatusOK {
				c.Status(StatusNoContent)
			}
			return nil
		}

		// Marshal and send the response
		if err := c.JSON(StatusOK, res); err != nil {
			c.AbortWithError(StatusInternalServerError, NewHTTPError(
				StatusInternalServerError,
				fmt.Sprintf("Failed to marshal response: %v", err),
			))
			return nil
		}
		return nil
	}
}

// getTypeName returns a human-readable type name.
func getTypeName(t reflect.Type) string {
	if t == nil {
		return "nil"
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		return "*" + getTypeName(t.Elem())
	}

	// Use package path + name for named types
	if t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}

	// For anonymous types, use Kind
	return t.String()
}

// bindTypedRequest attempts to bind the request from various sources.
// It tries to bind from the request body, query params, and path params.
func bindTypedRequest(c *Context, v any) error {
	method := c.Req.Method
	contentType := c.Header("Content-Type")

	// For GET, DELETE, and HEAD requests, bind from query and path params
	if method == "GET" || method == "DELETE" || method == "HEAD" {
		// First bind path parameters
		if len(c.Params) > 0 {
			if err := bindParams(c.Params, v); err != nil {
				return &BindingError{
					Source: "path parameters",
					Cause:  err,
				}
			}
		}

		// Then bind query parameters
		if len(c.Req.URL.Query()) > 0 {
			if err := bindMap(c.Req.URL.Query(), v, "query"); err != nil {
				return &BindingError{
					Source: "query parameters",
					Cause:  err,
				}
			}
		}

		return nil
	}

	// For POST, PUT, PATCH, try to bind from body
	if method == "POST" || method == "PUT" || method == "PATCH" {
		// First bind path parameters if any
		if len(c.Params) > 0 {
			if err := bindParams(c.Params, v); err != nil {
				return &BindingError{
					Source: "path parameters",
					Cause:  err,
				}
			}
		}

		// Then bind body based on content type
		switch contentType {
		case "", "application/json":
			if c.Req.Body != nil {
				if err := json.NewDecoder(c.Req.Body).Decode(v); err != nil {
					return &BindingError{
						Source:      "JSON body",
						Cause:       err,
						ContentType: contentType,
					}
				}
			}
		case "application/x-www-form-urlencoded", "multipart/form-data":
			if err := bindForm(c.Req, v); err != nil {
				return &BindingError{
					Source:      "form data",
					Cause:       err,
					ContentType: contentType,
				}
			}
		default:
			return &BindingError{
				Source:      "request body",
				ContentType: contentType,
				Cause:       fmt.Errorf("unsupported content type: %s", contentType),
			}
		}

		return nil
	}

	return nil
}

// BindingError provides detailed information about binding failures.
type BindingError struct {
	Source      string // e.g., "JSON body", "query parameters"
	ContentType string // Content-Type header if relevant
	Cause       error  // Underlying error
}

func (e *BindingError) Error() string {
	if e.ContentType != "" {
		return fmt.Sprintf("binding from %s (Content-Type: %s) failed: %v", e.Source, e.ContentType, e.Cause)
	}
	return fmt.Sprintf("binding from %s failed: %v", e.Source, e.Cause)
}

func (e *BindingError) Unwrap() error {
	return e.Cause
}

// TypedHandlerWithStatus is like TypedHandler but also returns an HTTP status code.
type TypedHandlerWithStatus[Req any, Res any] func(*Context, Req) (int, Res, error)

// TypedHandlerWithStatusFunc wraps a typed handler with custom status code.
func TypedHandlerWithStatusFunc[Req any, Res any](handler TypedHandlerWithStatus[Req, Res]) Handler {
	// Cache type checks at handler creation time
	var emptyReq Req
	var emptyRes Res

	reqType := reflect.TypeOf(emptyReq)
	resType := reflect.TypeOf(emptyRes)
	emptyReqType := reflect.TypeOf(EmptyRequest{})

	isEmptyReq := reqType == emptyReqType
	isEmptyRes := resType == emptyReqType

	reqTypeName := getTypeName(reqType)

	return func(c *Context) error {
		var req Req

		// Check if Req is EmptyRequest
		if !isEmptyReq {
			if err := bindTypedRequest(c, &req); err != nil {
				errorMsg := fmt.Sprintf("Failed to bind request to type %s: %v", reqTypeName, err)
				c.AbortWithError(StatusBadRequest, NewHTTPError(StatusBadRequest, errorMsg))
				return nil
			}

			if err := validateStruct(req); err != nil {
				c.AbortWithError(StatusUnprocessableEntity, err)
				return nil
			}
		}

		// Call the typed handler with status
		status, res, err := handler(c, req)
		if err != nil {
			if httpErr, ok := err.(*HTTPError); ok {
				c.AbortWithError(httpErr.Code, httpErr)
			} else {
				c.AbortWithError(StatusInternalServerError, err)
			}
			return nil
		}

		// Check if Res is EmptyRequest
		if isEmptyRes {
			c.Status(status)
			return nil
		}

		// Send response with custom status
		if err := c.JSON(status, res); err != nil {
			c.AbortWithError(StatusInternalServerError, NewHTTPError(
				StatusInternalServerError,
				fmt.Sprintf("Failed to marshal response: %v", err),
			))
			return nil
		}
		return nil
	}
}

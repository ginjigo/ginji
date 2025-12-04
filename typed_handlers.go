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
	return func(c *Context) {
		var req Req

		// Check if Req is EmptyRequest, in which case we skip binding
		reqType := reflect.TypeOf((*Req)(nil)).Elem()
		if reqType != reflect.TypeOf(EmptyRequest{}) {
			// Attempt to bind the request
			if err := bindTypedRequest(c, &req); err != nil {
				c.AbortWithError(StatusBadRequest, NewHTTPError(StatusBadRequest, "Invalid request: "+err.Error()))
				return
			}

			// Validate the bound request
			if err := validateStruct(req); err != nil {
				c.AbortWithError(StatusBadRequest, err)
				return
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
			return
		}

		// Check if Res is EmptyRequest (used to represent no response body)
		resType := reflect.TypeOf((*Res)(nil)).Elem()
		if resType == reflect.TypeOf(EmptyRequest{}) {
			// No response body, just return 204 No Content or keep existing status
			if c.StatusCode() == StatusOK {
				c.Status(StatusNoContent)
			}
			return
		}

		// Marshal and send the response
		if err := c.JSON(StatusOK, res); err != nil {
			c.AbortWithError(StatusInternalServerError, err)
		}
	}
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
				return fmt.Errorf("path binding failed: %w", err)
			}
		}

		// Then bind query parameters
		if len(c.Req.URL.Query()) > 0 {
			if err := bindMap(c.Req.URL.Query(), v, "query"); err != nil {
				return fmt.Errorf("query binding failed: %w", err)
			}
		}

		return nil
	}

	// For POST, PUT, PATCH, try to bind from body
	if method == "POST" || method == "PUT" || method == "PATCH" {
		// First bind path parameters if any
		if len(c.Params) > 0 {
			if err := bindParams(c.Params, v); err != nil {
				return fmt.Errorf("path binding failed: %w", err)
			}
		}

		// Then bind body based on content type
		switch contentType {
case "", "application/json":
			if c.Req.Body != nil {
				if err := json.NewDecoder(c.Req.Body).Decode(v); err != nil {
					return fmt.Errorf("json binding failed: %w", err)
				}
			}
		case "application/x-www-form-urlencoded", "multipart/form-data":
			if err := bindForm(c.Req, v); err != nil {
				return fmt.Errorf("form binding failed: %w", err)
			}
		}

		return nil
	}

	return nil
}

// TypedHandlerWithStatus is like TypedHandler but also returns an HTTP status code.
type TypedHandlerWithStatus[Req any, Res any] func(*Context, Req) (int, Res, error)

// TypedHandlerWithStatusFunc wraps a typed handler with custom status code.
func TypedHandlerWithStatusFunc[Req any, Res any](handler TypedHandlerWithStatus[Req, Res]) Handler {
	return func(c *Context) {
		var req Req

		// Check if Req is EmptyRequest
		reqType := reflect.TypeOf((*Req)(nil)).Elem()
		if reqType != reflect.TypeOf(EmptyRequest{}) {
			if err := bindTypedRequest(c, &req); err != nil {
				c.AbortWithError(StatusBadRequest, NewHTTPError(StatusBadRequest, "Invalid request: "+err.Error()))
				return
			}

			if err := validateStruct(req); err != nil {
				c.AbortWithError(StatusBadRequest, err)
				return
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
			return
		}

		// Check if Res is EmptyRequest
		resType := reflect.TypeOf((*Res)(nil)).Elem()
		if resType == reflect.TypeOf(EmptyRequest{}) {
			c.Status(status)
			return
		}

		// Send response with custom status
		if err := c.JSON(status, res); err != nil {
			c.AbortWithError(StatusInternalServerError, err)
		}
	}
}

package ginji

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNewTestContext(t *testing.T) {
	c, w := NewTestContext()

	if c == nil {
		t.Fatal("Expected context to be created")
	}
	if w == nil {
		t.Fatal("Expected response recorder to be created")
	}

	if c.Req.Method != "GET" {
		t.Errorf("Expected method GET, got %s", c.Req.Method)
	}
}

func TestPerformRequest(t *testing.T) {
	app := New()
	app.Get("/test", func(c *Context) {
		_ = c.Text(StatusOK, "Hello Test")
	})

	w := PerformRequest(app, "GET", "/test", nil)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello Test" {
		t.Errorf("Expected body 'Hello Test', got '%s'", w.Body.String())
	}
}

func TestPerformJSONRequest(t *testing.T) {
	app := New()
	app.Post("/api/user", func(c *Context) {
		var data map[string]string
		if err := c.BindJSON(&data); err != nil {
			_ = c.JSON(StatusBadRequest, H{"error": err.Error()})
			return
		}
		_ = c.JSON(StatusOK, H{"received": data["name"]})
	})

	payload := map[string]string{"name": "John"}
	w := PerformJSONRequest(app, "POST", "/api/user", payload)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestPerformFormRequest(t *testing.T) {
	app := New()
	app.Post("/submit", func(c *Context) {
		name := c.FormValue("name")
		_ = c.JSON(StatusOK, H{"name": name})
	})

	formData := url.Values{}
	formData.Set("name", "Alice")

	w := PerformFormRequest(app, "POST", "/submit", formData)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestBuilder(t *testing.T) {
	app := New()
	app.Get("/hello", func(c *Context) {
		auth := c.Header("Authorization")
		_ = c.JSON(StatusOK, H{"auth": auth})
	})

	w := NewRequest(app, "GET", "/hello").
		Header("Authorization", "Bearer token123").
		Do()

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestBuilderJSON(t *testing.T) {
	app := New()
	app.Post("/data", func(c *Context) {
		var data H
		if err := c.BindJSON(&data); err != nil {
			_ = c.JSON(StatusBadRequest, H{"error": err.Error()})
			return
		}
		_ = c.JSON(StatusOK, data)
	})

	w := NewRequest(app, "POST", "/data").
		JSON(H{"key": "value"}).
		Do()

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRequestBuilderForm(t *testing.T) {
	app := New()
	app.Post("/form", func(c *Context) {
		email := c.FormValue("email")
		_ = c.Text(StatusOK, email)
	})

	formData := url.Values{}
	formData.Set("email", "test@example.com")

	w := NewRequest(app, "POST", "/form").
		Form(formData).
		Do()

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "test@example.com" {
		t.Errorf("Expected body 'test@example.com', got '%s'", w.Body.String())
	}
}

func TestRequestBuilderCookie(t *testing.T) {
	app := New()
	app.Get("/cookie-test", func(c *Context) {
		cookie, err := c.Cookie("session")
		if err != nil {
			_ = c.Text(StatusBadRequest, "no cookie")
			return
		}
		_ = c.Text(StatusOK, cookie.Value)
	})

	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
	}

	w := NewRequest(app, "GET", "/cookie-test").
		Cookie(cookie).
		Do()

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "abc123" {
		t.Errorf("Expected body 'abc123', got '%s'", w.Body.String())
	}
}

func TestResponseHelpers(t *testing.T) {
	app := New()
	app.Get("/json", func(c *Context) {
		_ = c.JSON(StatusOK, H{"message": "hello"})
	})

	w := NewRequest(app, "GET", "/json").Do()
	resp := NewResponse(w)

	if resp.Status() != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Status())
	}

	var data H
	if err := resp.JSON(&data); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if data["message"] != "hello" {
		t.Errorf("Expected message 'hello', got '%v'", data["message"])
	}
}

func TestMockMiddleware(t *testing.T) {
	app := New()
	app.Use(MockMiddleware("test_key", "test_value"))
	app.Get("/test", func(c *Context) {
		val := c.GetString("test_key")
		_ = c.Text(StatusOK, val)
	})

	w := PerformRequest(app, "GET", "/test", nil)

	if w.Body.String() != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", w.Body.String())
	}
}

func TestAssertHelpers(t *testing.T) {
	app := New()
	app.Get("/status-test", func(c *Context) {
		_ = c.JSON(StatusCreated, H{"status": "created"})
	})

	w := PerformRequest(app, "GET", "/status-test", nil)

	// Test assertion helpers
	AssertStatus(t, w, StatusCreated)
	AssertBody(t, w, "created")
	AssertHeader(t, w, "Content-Type", "application/json")
}

func TestPerformMultipartRequest(t *testing.T) {
	app := New()
	app.Post("/upload", func(c *Context) {
		name := c.FormValue("name")
		file, err := c.FormFile("file")
		if err != nil {
			_ = c.Text(StatusBadRequest, "no file")
			return
		}
		_ = c.JSON(StatusOK, H{"name": name, "filename": file.Filename})
	})

	fields := map[string]string{"name": "test_upload"}
	files := map[string][]byte{"file": []byte("file content")}

	w := PerformMultipartRequest(app, "POST", "/upload", fields, files)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

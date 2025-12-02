package ginji

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type TestStruct struct {
	Name string `query:"name" header:"X-Name"`
}

func TestBindQuery(t *testing.T) {
	app := New()
	app.Get("/test", func(c *Context) {
		var ts TestStruct
		if err := c.BindQuery(&ts); err != nil {
			c.Text(http.StatusInternalServerError, err.Error())
			return
		}
		c.Text(http.StatusOK, ts.Name)
	})

	req := httptest.NewRequest("GET", "/test?name=ginji", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Body.String() != "ginji" {
		t.Errorf("Expected ginji, got %s", w.Body.String())
	}
}

func TestBindHeader(t *testing.T) {
	app := New()
	app.Get("/test", func(c *Context) {
		var ts TestStruct
		if err := c.BindHeader(&ts); err != nil {
			c.Text(http.StatusInternalServerError, err.Error())
			return
		}
		c.Text(http.StatusOK, ts.Name)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Name", "ginji-header")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Body.String() != "ginji-header" {
		t.Errorf("Expected ginji-header, got %s", w.Body.String())
	}
}

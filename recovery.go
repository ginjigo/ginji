package ginji

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
)

// Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(c *Context) {
			defer func() {
				if err := recover(); err != nil {
					message := fmt.Sprintf("%s", err)
					log.Printf("%s\n\n", trace(message))
					c.Text(http.StatusInternalServerError, "Internal Server Error")
				}
			}()
			next(c)
		}
	}
}

// trace returns a stack trace for the panic.
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 frames
	var str string
	str += message + "\nTraceback:"
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str += fmt.Sprintf("\n\t%s:%d", file, line)
	}
	return str
}

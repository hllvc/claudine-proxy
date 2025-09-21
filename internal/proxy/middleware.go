package proxy

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v3"
)

// Recovery recovers from panics in HTTP handlers and returns HTTP 500 to the client.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recover() != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				// Logging of panics is handled in Logging middleware
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Logging logs HTTP requests with method, path, status, and duration.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return httplog.RequestLogger(logger, &httplog.Options{
		Schema: httplog.SchemaECS.Concise(true),

		// Explicitly prevent logging headers/body to avoid leaking sensitive data
		LogRequestHeaders:  []string{"Content-Type", "Origin"}, // Default, but explicit
		LogResponseHeaders: []string{},                         // Explicit empty (default is empty, but be clear)
		LogRequestBody:     nil,                                // Never log request bodies (default, but explicit)
		LogResponseBody:    nil,                                // Never log response bodies (default, but explicit)

		RecoverPanics: false, // use dedicated middleware, panics are logged regardless
	})
}

// applyMiddlewares applies middlewares to a handler in the order they appear.
// The first middleware in the slice is the outermost (executes first).
func applyMiddlewares(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

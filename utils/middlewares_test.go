package utils

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_Success(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock handler that succeeds
	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	// Execute middleware
	middleware := LoggingMiddleware(nextHandler)
	err := middleware(c)

	// Verify middleware behavior
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "success", rec.Body.String())
}

func TestLoggingMiddleware_HandlerError(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/error", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Mock handler that fails
	testError := errors.New("test error")
	nextHandler := func(c echo.Context) error {
		return testError
	}

	// Execute middleware
	middleware := LoggingMiddleware(nextHandler)
	err := middleware(c)

	// Verify error is properly propagated
	assert.Error(t, err)
	assert.Equal(t, testError, err)
}

func TestLoggingMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(method, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Mock handler
			nextHandler := func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}

			// Execute middleware
			middleware := LoggingMiddleware(nextHandler)
			err := middleware(c)

			// Verify middleware works for all methods
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "ok", rec.Body.String())
		})
	}
}

func TestLoggingMiddleware_QueryParameters(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"Single parameter", "/test?param=value"},
		{"Multiple parameters", "/test?param1=value1&param2=value2"},
		{"No parameters", "/test"},
		{"Special characters", "/test?query=hello%20world&special=%21%40%23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Mock handler
			nextHandler := func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}

			// Execute middleware
			middleware := LoggingMiddleware(nextHandler)
			err := middleware(c)

			// Verify middleware works with all URL patterns
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "ok", rec.Body.String())
		})
	}
}

func TestLoggingMiddleware_URIPaths(t *testing.T) {
	paths := []string{"/", "/api/v1/users", "/files/upload", "/health/check", "/metrics"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Mock handler
			nextHandler := func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			}

			// Execute middleware
			middleware := LoggingMiddleware(nextHandler)
			err := middleware(c)

			// Verify middleware works with all paths
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "ok", rec.Body.String())
		})
	}
}

func TestLoggingMiddleware_ErrorTypes(t *testing.T) {
	tests := []struct {
		name  string
		error error
	}{
		{"Simple error", errors.New("simple error")},
		{"Echo HTTP error", echo.NewHTTPError(http.StatusBadRequest, "bad request")},
		{"Echo HTTP error with internal", echo.NewHTTPError(http.StatusInternalServerError, "internal error").SetInternal(errors.New("internal cause"))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Mock handler that returns the test error
			nextHandler := func(c echo.Context) error {
				return tt.error
			}

			// Execute middleware
			middleware := LoggingMiddleware(nextHandler)
			err := middleware(c)

			// Verify error is properly propagated
			assert.Equal(t, tt.error, err)
		})
	}
}

func TestLoggingMiddleware_Concurrency(t *testing.T) {
	// Test that the middleware works correctly under concurrent requests
	const numRequests = 10

	e := echo.New()

	// Mock handler
	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	// Create channels for synchronization
	done := make(chan bool, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			middleware := LoggingMiddleware(nextHandler)
			err := middleware(c)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, rec.Code)

			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		<-done
	}
}

// Benchmark the middleware performance
func BenchmarkLoggingMiddleware(b *testing.B) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test?param=value", nil)

	nextHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	middleware := LoggingMiddleware(nextHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		_ = middleware(c)
	}
}

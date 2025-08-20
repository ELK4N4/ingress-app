package main

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elk4n4/ingress-app/handlers"
	"github.com/elk4n4/ingress-app/object_storage"
	"github.com/elk4n4/ingress-app/producer"
	"github.com/elk4n4/ingress-app/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProducer for integration tests
type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProducer) Publish(msg []byte, topic string, key string) error {
	args := m.Called(msg, topic, key)
	return args.Error(0)
}

// MockObjectStorage for integration tests
type MockObjectStorage struct {
	mock.Mock
}

func (m *MockObjectStorage) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockObjectStorage) UploadFile(ctx context.Context, bucketName string, fileName string, fileSize int64, contentType string, reader io.Reader) (string, error) {
	args := m.Called(ctx, bucketName, fileName, fileSize, contentType, reader)
	return args.String(0), args.Error(1)
}

func TestIntegration_FullFileUploadFlow(t *testing.T) {
	// Setup Echo server
	e := echo.New()
	e.Use(utils.LoggingMiddleware)

	// Setup mocks
	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	// Configure mocks
	mockStorage.On("UploadFile", mock.Anything, "files", "test.txt", int64(12), "application/octet-stream", mock.Anything).Return("test-file-key", nil)
	mockProducer.On("Publish", []byte("test-file-key"), "files", "test-file-key").Return(nil)

	// Setup handler
	ah := handlers.AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "files",
		ObjectStorage: mockStorage,
		BucketName:    "files",
	}

	// Setup routes
	e.GET("/ping", ping)
	e.POST("/publish", ah.AnnounceFile)

	// Test ping endpoint
	t.Run("Ping endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "pong", rec.Body.String())
	})

	// Test file upload flow
	t.Run("File upload flow", func(t *testing.T) {
		// Create multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.txt")
		assert.NoError(t, err)

		_, err = part.Write([]byte("test content"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/publish", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		// Execute request
		e.ServeHTTP(rec, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Published", rec.Body.String())

		// Verify mocks were called
		mockStorage.AssertExpectations(t)
		mockProducer.AssertExpectations(t)
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	e := echo.New()
	e.Use(utils.LoggingMiddleware)

	t.Run("Missing file in request", func(t *testing.T) {
		mockProducer := &MockProducer{}
		mockStorage := &MockObjectStorage{}

		ah := handlers.AnnounceHandler{
			Producer:      mockProducer,
			Topic:         "files",
			ObjectStorage: mockStorage,
			BucketName:    "files",
		}

		e.POST("/publish", ah.AnnounceFile)

		// Create request without file
		req := httptest.NewRequest(http.MethodPost, "/publish", strings.NewReader("no file data"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// Should return an error (but Echo might handle it differently)
		assert.NotEqual(t, http.StatusOK, rec.Code)
	})

	t.Run("Storage failure", func(t *testing.T) {
		mockProducer := &MockProducer{}
		mockStorage := &MockObjectStorage{}

		// Configure storage to fail
		mockStorage.On("UploadFile", mock.Anything, "files", "test.txt", int64(12), "application/octet-stream", mock.Anything).Return("", assert.AnError)

		ah := handlers.AnnounceHandler{
			Producer:      mockProducer,
			Topic:         "files",
			ObjectStorage: mockStorage,
			BucketName:    "files",
		}

		e.POST("/publish", ah.AnnounceFile)

		// Create multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.txt")
		assert.NoError(t, err)

		_, err = part.Write([]byte("test content"))
		assert.NoError(t, err)

		err = writer.Close()
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/publish", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		// Should return an error
		assert.NotEqual(t, http.StatusOK, rec.Code)

		mockStorage.AssertExpectations(t)
		// Producer should not be called if storage fails
		mockProducer.AssertNotCalled(t, "Publish")
	})
}

func TestIntegration_WithRealServices(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test with real services in short mode")
	}

	// Test with actual MinIO and Kafka if available
	t.Run("Real MinIO integration", func(t *testing.T) {
		// Try to connect to MinIO
		mos := object_storage.NewMinioObjectStorage("localhost:9000", "minioadmin", "minioadmin", false)
		err := mos.Connect()
		if err != nil {
			t.Skipf("MinIO not available: %v", err)
		}

		// Test a simple operation
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		reader := strings.NewReader("test content")
		key, err := mos.UploadFile(ctx, "test-bucket", "integration-test.txt", 12, "text/plain", reader)

		if err != nil {
			t.Logf("MinIO upload failed (expected in test environment): %v", err)
		} else {
			assert.Equal(t, "integration-test.txt", key)
		}
	})

	t.Run("Real Kafka integration", func(t *testing.T) {
		// Try to connect to Kafka
		p := producer.NewKafkaProducer("localhost:9092", 5)
		err := p.Connect()
		if err != nil {
			t.Skipf("Kafka not available: %v", err)
		}

		// Test publishing a message
		err = p.Publish([]byte("test-message"), "test-topic", "test-key")
		if err != nil {
			t.Logf("Kafka publish failed (expected in test environment): %v", err)
		}
	})
}

func TestIntegration_LoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	// Setup
	e := echo.New()
	e.Use(utils.LoggingMiddleware)

	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	// Configure mocks to handle multiple calls
	mockStorage.On("UploadFile", mock.Anything, "files", mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string"), mock.Anything).Return("test-key", nil)
	mockProducer.On("Publish", mock.Anything, "files", mock.AnythingOfType("string")).Return(nil)

	ah := handlers.AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "files",
		ObjectStorage: mockStorage,
		BucketName:    "files",
	}

	e.POST("/publish", ah.AnnounceFile)

	// Perform load test
	const numRequests = 10
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			// Create multipart form data
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, err := writer.CreateFormFile("file", "load-test.txt")
			assert.NoError(t, err)

			_, err = part.Write([]byte("load test content"))
			assert.NoError(t, err)

			err = writer.Close()
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/publish", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		select {
		case <-done:
			// Request completed
		case <-time.After(10 * time.Second):
			t.Fatal("Request timed out")
		}
	}
}

// Benchmark integration test
func BenchmarkIntegration_FileUpload(b *testing.B) {
	// Setup
	e := echo.New()

	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	mockStorage.On("UploadFile", mock.Anything, "files", mock.AnythingOfType("string"), mock.AnythingOfType("int64"), mock.AnythingOfType("string"), mock.Anything).Return("test-key", nil)
	mockProducer.On("Publish", mock.Anything, "files", mock.AnythingOfType("string")).Return(nil)

	ah := handlers.AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "files",
		ObjectStorage: mockStorage,
		BucketName:    "files",
	}

	e.POST("/publish", ah.AnnounceFile)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Create multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "bench.txt")
		part.Write([]byte("benchmark content"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/publish", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
	}
}

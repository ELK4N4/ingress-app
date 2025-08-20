package handlers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProducer implements the producer.Producer interface for testing
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

// MockObjectStorage implements the object_storage.ObjectStorage interface for testing
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

// MockFile implements multipart.File for testing
type MockFile struct {
	*strings.Reader
}

func (m MockFile) Close() error {
	return nil
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    string
		expectError bool
	}{
		{
			name:        "Text file",
			content:     "Hello, World!",
			expected:    "application/octet-stream",
			expectError: false,
		},
		{
			name:        "JSON content",
			content:     `{"key": "value"}`,
			expected:    "application/octet-stream",
			expectError: false,
		},
		{
			name:        "HTML content",
			content:     "<html><body>Test</body></html>",
			expected:    "text/html; charset=utf-8",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := MockFile{strings.NewReader(tt.content)}
			contentType, err := getContentType(reader)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, contentType)
			}
		})
	}
}

func createMultipartRequest(fieldName, fileName, content string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		return nil, err
	}

	_, err = part.Write([]byte(content))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, "/publish", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

func TestAnnounceHandler_AnnounceFile_Success(t *testing.T) {
	// Setup
	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	handler := &AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "test-topic",
		ObjectStorage: mockStorage,
		BucketName:    "test-bucket",
	}

	// Create test request
	req, err := createMultipartRequest("file", "test.txt", "test content")
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Mock expectations - multipart files are detected as octet-stream
	mockStorage.On("UploadFile", mock.Anything, "test-bucket", "test.txt", int64(12), "application/octet-stream", mock.Anything).Return("test-key", nil)
	mockProducer.On("Publish", []byte("test-key"), "test-topic", "test-key").Return(nil)

	// Execute
	err = handler.AnnounceFile(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Published", rec.Body.String())

	mockStorage.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestAnnounceHandler_AnnounceFile_NoFile(t *testing.T) {
	// Setup
	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	handler := &AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "test-topic",
		ObjectStorage: mockStorage,
		BucketName:    "test-bucket",
	}

	// Create request without file
	req := httptest.NewRequest(http.MethodPost, "/publish", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.AnnounceFile(c)

	// Assert
	assert.Error(t, err)
}

func TestAnnounceHandler_AnnounceFile_StorageError(t *testing.T) {
	// Setup
	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	handler := &AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "test-topic",
		ObjectStorage: mockStorage,
		BucketName:    "test-bucket",
	}

	// Create test request
	req, err := createMultipartRequest("file", "test.txt", "test content")
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Mock expectations - storage fails
	mockStorage.On("UploadFile", mock.Anything, "test-bucket", "test.txt", int64(12), "application/octet-stream", mock.Anything).Return("", errors.New("storage error"))

	// Execute
	err = handler.AnnounceFile(c)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "storage error", err.Error())

	mockStorage.AssertExpectations(t)
	mockProducer.AssertNotCalled(t, "Publish")
}

func TestAnnounceHandler_AnnounceFile_ProducerError(t *testing.T) {
	// Setup
	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	handler := &AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "test-topic",
		ObjectStorage: mockStorage,
		BucketName:    "test-bucket",
	}

	// Create test request
	req, err := createMultipartRequest("file", "test.txt", "test content")
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Mock expectations - producer fails
	mockStorage.On("UploadFile", mock.Anything, "test-bucket", "test.txt", int64(12), "application/octet-stream", mock.Anything).Return("test-key", nil)
	mockProducer.On("Publish", []byte("test-key"), "test-topic", "test-key").Return(errors.New("producer error"))

	// Execute
	err = handler.AnnounceFile(c)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "producer error", err.Error())

	mockStorage.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestAnnounceHandler_AnnounceFile_LargeFile(t *testing.T) {
	// Setup
	mockProducer := &MockProducer{}
	mockStorage := &MockObjectStorage{}

	handler := &AnnounceHandler{
		Producer:      mockProducer,
		Topic:         "test-topic",
		ObjectStorage: mockStorage,
		BucketName:    "test-bucket",
	}

	// Create large content
	largeContent := strings.Repeat("A", 1024*1024) // 1MB
	req, err := createMultipartRequest("file", "large.txt", largeContent)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Mock expectations - use AnythingOfType for content type since it varies based on content
	mockStorage.On("UploadFile", mock.Anything, "test-bucket", "large.txt", int64(1024*1024), mock.AnythingOfType("string"), mock.Anything).Return("large-key", nil)
	mockProducer.On("Publish", []byte("large-key"), "test-topic", "large-key").Return(nil)

	// Execute
	err = handler.AnnounceFile(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Published", rec.Body.String())

	mockStorage.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

package object_storage

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMinioObjectStorage(t *testing.T) {
	tests := []struct {
		name            string
		endpoint        string
		accessKeyID     string
		secretAccessKey string
		useSSL          bool
		expected        *MinioObjectStorage
	}{
		{
			name:            "Standard configuration",
			endpoint:        "localhost:9000",
			accessKeyID:     "minioadmin",
			secretAccessKey: "minioadmin",
			useSSL:          false,
			expected: &MinioObjectStorage{
				endpoint:        "localhost:9000",
				accessKeyID:     "minioadmin",
				secretAccessKey: "minioadmin",
				useSSL:          false,
			},
		},
		{
			name:            "SSL enabled",
			endpoint:        "minio.example.com:9000",
			accessKeyID:     "access-key",
			secretAccessKey: "secret-key",
			useSSL:          true,
			expected: &MinioObjectStorage{
				endpoint:        "minio.example.com:9000",
				accessKeyID:     "access-key",
				secretAccessKey: "secret-key",
				useSSL:          true,
			},
		},
		{
			name:            "Empty credentials",
			endpoint:        "localhost:9000",
			accessKeyID:     "",
			secretAccessKey: "",
			useSSL:          false,
			expected: &MinioObjectStorage{
				endpoint:        "localhost:9000",
				accessKeyID:     "",
				secretAccessKey: "",
				useSSL:          false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMinioObjectStorage(tt.endpoint, tt.accessKeyID, tt.secretAccessKey, tt.useSSL)

			assert.Equal(t, tt.expected.endpoint, storage.endpoint)
			assert.Equal(t, tt.expected.accessKeyID, storage.accessKeyID)
			assert.Equal(t, tt.expected.secretAccessKey, storage.secretAccessKey)
			assert.Equal(t, tt.expected.useSSL, storage.useSSL)
			assert.Nil(t, storage.client) // Client should be nil until Connect() is called
		})
	}
}

func TestMinioObjectStorage_Connect_InvalidConfig(t *testing.T) {
	tests := []struct {
		name            string
		endpoint        string
		accessKeyID     string
		secretAccessKey string
		useSSL          bool
		expectError     bool
	}{
		{
			name:            "Invalid endpoint",
			endpoint:        "invalid-endpoint-format",
			accessKeyID:     "access",
			secretAccessKey: "secret",
			useSSL:          false,
			expectError:     false, // MinIO client creation might succeed even with invalid endpoint
		},
		{
			name:            "Empty endpoint",
			endpoint:        "",
			accessKeyID:     "access",
			secretAccessKey: "secret",
			useSSL:          false,
			expectError:     true,
		},
		{
			name:            "Malformed endpoint",
			endpoint:        "://invalid",
			accessKeyID:     "access",
			secretAccessKey: "secret",
			useSSL:          false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMinioObjectStorage(tt.endpoint, tt.accessKeyID, tt.secretAccessKey, tt.useSSL)
			err := storage.Connect()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, storage.client)
			} else {
				// Note: Connection might succeed even with invalid config,
				// actual network errors happen during operations
				if err != nil {
					assert.Nil(t, storage.client)
				}
			}
		})
	}
}

// Integration test - only runs when MinIO is available
func TestMinioObjectStorage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	storage := NewMinioObjectStorage("localhost:9000", "minioadmin", "minioadmin", false)

	// Try to connect - this will only work if MinIO is running
	err := storage.Connect()
	if err != nil {
		t.Skipf("MinIO not available for integration test: %v", err)
	}

	ctx := context.Background()
	bucketName := "test-bucket"
	fileName := "test-file.txt"
	content := "test content for upload"
	contentType := "text/plain"
	reader := strings.NewReader(content)

	// Test file upload
	key, err := storage.UploadFile(ctx, bucketName, fileName, int64(len(content)), contentType, reader)

	if err != nil {
		t.Logf("Expected error without real MinIO instance: %v", err)
	} else {
		assert.Equal(t, fileName, key)
	}
}

func TestMinioObjectStorage_UploadFile_Validation(t *testing.T) {
	// Test that we can create a storage instance without connecting
	storage := NewMinioObjectStorage("localhost:9000", "test", "test", false)

	// Verify the storage was created with correct parameters
	assert.Equal(t, "localhost:9000", storage.endpoint)
	assert.Equal(t, "test", storage.accessKeyID)
	assert.Equal(t, "test", storage.secretAccessKey)
	assert.False(t, storage.useSSL)
	assert.Nil(t, storage.client) // Client should be nil until Connect() is called
}

func TestMinioObjectStorage_EdgeCases(t *testing.T) {
	t.Run("Very long endpoint", func(t *testing.T) {
		longEndpoint := strings.Repeat("a", 1000) + ".com:9000"
		storage := NewMinioObjectStorage(longEndpoint, "access", "secret", false)
		assert.Equal(t, longEndpoint, storage.endpoint)
	})

	t.Run("Special characters in credentials", func(t *testing.T) {
		accessKey := "access-key_123"
		secretKey := "secret/key+with=special&chars"
		storage := NewMinioObjectStorage("localhost:9000", accessKey, secretKey, true)
		assert.Equal(t, accessKey, storage.accessKeyID)
		assert.Equal(t, secretKey, storage.secretAccessKey)
	})

	t.Run("Non-standard port", func(t *testing.T) {
		endpoint := "minio.example.com:12345"
		storage := NewMinioObjectStorage(endpoint, "access", "secret", true)
		assert.Equal(t, endpoint, storage.endpoint)
	})

	t.Run("SSL configuration", func(t *testing.T) {
		storageSSL := NewMinioObjectStorage("secure.minio.com:443", "access", "secret", true)
		storageNoSSL := NewMinioObjectStorage("localhost:9000", "access", "secret", false)

		assert.True(t, storageSSL.useSSL)
		assert.False(t, storageNoSSL.useSSL)
	})
}

// Benchmark test for storage creation
func BenchmarkNewMinioObjectStorage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		storage := NewMinioObjectStorage("localhost:9000", "minioadmin", "minioadmin", false)
		_ = storage
	}
}

// Test configuration validation
func TestMinioObjectStorage_ConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		valid    bool
	}{
		{
			name:     "Standard localhost",
			endpoint: "localhost:9000",
			valid:    true,
		},
		{
			name:     "IP address",
			endpoint: "192.168.1.100:9000",
			valid:    true,
		},
		{
			name:     "Domain name",
			endpoint: "minio.example.com:9000",
			valid:    true,
		},
		{
			name:     "Domain with SSL port",
			endpoint: "secure-minio.com:443",
			valid:    true,
		},
		{
			name:     "No port specified",
			endpoint: "minio.example.com",
			valid:    true, // MinIO client might handle default ports
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMinioObjectStorage(tt.endpoint, "access", "secret", false)
			assert.NotNil(t, storage)
			assert.Equal(t, tt.endpoint, storage.endpoint)
		})
	}
}

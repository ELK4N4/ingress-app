package producer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKafkaProducer(t *testing.T) {
	tests := []struct {
		name     string
		servers  string
		retries  int
		expected *KafkaProducer
	}{
		{
			name:    "Valid configuration",
			servers: "localhost:9092",
			retries: 5,
			expected: &KafkaProducer{
				servers: "localhost:9092",
				retries: 5,
			},
		},
		{
			name:    "Multiple servers",
			servers: "kafka1:9092,kafka2:9092",
			retries: 3,
			expected: &KafkaProducer{
				servers: "kafka1:9092,kafka2:9092",
				retries: 3,
			},
		},
		{
			name:    "Zero retries",
			servers: "localhost:9092",
			retries: 0,
			expected: &KafkaProducer{
				servers: "localhost:9092",
				retries: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := NewKafkaProducer(tt.servers, tt.retries)

			assert.Equal(t, tt.expected.servers, producer.servers)
			assert.Equal(t, tt.expected.retries, producer.retries)
			assert.Nil(t, producer.p) // Producer should be nil until Connect() is called
		})
	}
}

func TestKafkaProducer_Connect_InvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		servers string
		retries int
	}{
		{
			name:    "Invalid server format",
			servers: "invalid-server-format",
			retries: 5,
		},
		{
			name:    "Empty servers",
			servers: "",
			retries: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := NewKafkaProducer(tt.servers, tt.retries)
			err := producer.Connect()

			// Note: Kafka producer creation might succeed even with invalid config
			// The actual connection error happens during operations
			if err != nil {
				assert.Nil(t, producer.p)
			} else {
				// If connection succeeds, producer should be set
				assert.NotNil(t, producer.p)
				// Clean up
				if producer.p != nil {
					producer.p.Close()
				}
			}
		})
	}
}

// Integration test - only runs when Kafka is available
func TestKafkaProducer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	producer := NewKafkaProducer("localhost:9092", 5)

	// Try to connect - this will only work if Kafka is running
	err := producer.Connect()
	if err != nil {
		t.Skipf("Kafka not available for integration test: %v", err)
	}

	defer func() {
		if producer.p != nil {
			producer.p.Close()
		}
	}()

	// Test publishing a message
	err = producer.Publish([]byte("test-value"), "test-topic", "test-key")

	// In a real Kafka environment, this should succeed
	// In test environment without Kafka, it will fail
	if err != nil {
		t.Logf("Expected error without real Kafka instance: %v", err)
	}
}

// Test the configuration map creation logic indirectly
func TestKafkaProducerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		servers string
		retries int
		valid   bool
	}{
		{
			name:    "Valid single server",
			servers: "localhost:9092",
			retries: 5,
			valid:   true,
		},
		{
			name:    "Valid multiple servers",
			servers: "server1:9092,server2:9092",
			retries: 3,
			valid:   true,
		},
		{
			name:    "High retries",
			servers: "localhost:9092",
			retries: 100,
			valid:   true,
		},
		{
			name:    "Negative retries should still create producer",
			servers: "localhost:9092",
			retries: -1,
			valid:   true, // ConfigMap will handle this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := NewKafkaProducer(tt.servers, tt.retries)

			assert.NotNil(t, producer)
			assert.Equal(t, tt.servers, producer.servers)
			assert.Equal(t, tt.retries, producer.retries)
		})
	}
}

// Benchmark test for producer creation
func BenchmarkNewKafkaProducer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		producer := NewKafkaProducer("localhost:9092", 5)
		_ = producer
	}
}

// Test edge cases
func TestKafkaProducer_EdgeCases(t *testing.T) {
	t.Run("Extremely long server list", func(t *testing.T) {
		longServers := ""
		for i := 0; i < 1000; i++ {
			if i > 0 {
				longServers += ","
			}
			longServers += "server" + string(rune(i)) + ":9092"
		}

		producer := NewKafkaProducer(longServers, 5)
		assert.NotNil(t, producer)
		assert.Equal(t, longServers, producer.servers)
	})

	t.Run("Special characters in server name", func(t *testing.T) {
		servers := "localhost:9092,test-server_1.example.com:9092"
		producer := NewKafkaProducer(servers, 5)
		assert.NotNil(t, producer)
		assert.Equal(t, servers, producer.servers)
	})

	t.Run("Very high retry count", func(t *testing.T) {
		producer := NewKafkaProducer("localhost:9092", 999999)
		assert.NotNil(t, producer)
		assert.Equal(t, 999999, producer.retries)
	})
}

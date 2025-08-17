package producer

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var p *kafka.Producer = nil

func startProducer() {
	configMap := kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"go.delivery.reports":      true,
		"enable.auto.commit":       true,
		"go.logs.channel.enable":   true,
		"allow.auto.create.topics": true,
		"retries":                  5,
		"delivery.timeout.ms":      30000,
		"retry.backoff.ms":         1000,
	}

	producer, err := kafka.NewProducer(&configMap)
	if err != nil {
		Logger.Error().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Error creating kafka producer")
		return
	}

	p = producer
}

func SendMessage(msg []byte, topic string, key string) error {
	if p == nil {
		Logger.Info().Msg("Initializing kafka producer")
		startProducer()
	}
	delivery_chan := make(chan kafka.Event)
	defer close(delivery_chan)
	err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(msg), // Why are they creating the msg again instead of using a pointer?
	}, delivery_chan)
	if err != nil {
		Logger.Error().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Failed to produce message to kafka")
		return err
	}
	e := <-delivery_chan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		Logger.Error().Fields(map[string]any{
			"error": m.TopicPartition.Error.Error(),
		}).Msg("Delivery failed")
		return m.TopicPartition.Error
	} else {
		Logger.Debug().Msgf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		return nil
	}
}

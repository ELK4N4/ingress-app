package producer

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/elk4n4/ingress-app/utils"
)

type KafkaProducer struct {
	p       *kafka.Producer
	servers string
	retries int
}

func NewKafkaProducer(servers string, retries int) *KafkaProducer {
	return &KafkaProducer{servers: servers, retries: retries}
}

func (kp *KafkaProducer) Connect() error {
	configMap := kafka.ConfigMap{
		"bootstrap.servers":        kp.servers,
		"go.delivery.reports":      true,
		"enable.auto.commit":       true,
		"go.logs.channel.enable":   true,
		"allow.auto.create.topics": true,
		"retries":                  kp.retries,
		"delivery.timeout.ms":      30000,
		"retry.backoff.ms":         1000,
	}

	producer, err := kafka.NewProducer(&configMap)
	if err != nil {
		utils.Logger.Error().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Error creating kafka producer")
		return err
	}

	kp.p = producer
	return nil
}

func (kp *KafkaProducer) Publish(msg []byte, topic string, key string) error {
	delivery_chan := make(chan kafka.Event)
	defer close(delivery_chan)
	err := kp.p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(msg), // Why are they creating the msg again instead of using a pointer?
	}, delivery_chan)
	if err != nil {
		utils.Logger.Error().Fields(map[string]any{
			"error": err.Error(),
		}).Msg("Failed to produce message to kafka")
		return err
	}
	e := <-delivery_chan
	m := e.(*kafka.Message)
	if m.TopicPartition.Error != nil {
		utils.Logger.Error().Fields(map[string]any{
			"error": m.TopicPartition.Error.Error(),
		}).Msg("Delivery failed")
		return m.TopicPartition.Error
	} else {
		utils.Logger.Debug().Msgf("Delivered message to topic %s [%d] at offset %v\n",
			*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
		return nil
	}
}

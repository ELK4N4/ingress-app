package producer

type Producer interface {
	Connect() error
	Publish(msg []byte, topic string, key string) error
}

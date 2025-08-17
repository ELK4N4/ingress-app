package producer

type Producer interface {
	Publish(msg []byte, topic string, key string) error
}

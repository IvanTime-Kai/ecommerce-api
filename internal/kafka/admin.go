package kafka

import (
	"context"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func CreateTopic(ctx context.Context, brokerAddress, topic string, numPartitions int) error {

	conn, err := kafka.Dial("tcp", brokerAddress)

	if err != nil {
		return err
	}

	defer conn.Close()

	controller, err := conn.Controller()

	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))

	if err != nil {
		return err
	}
	defer controllerConn.Close()

	return controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     numPartitions,
		ReplicationFactor: 1,
	})
}

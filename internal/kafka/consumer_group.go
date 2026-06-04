package kafka

import (
	"context"
)

type ConsumerGroup struct {
	consumer []*Consumer
}

func NewConsumerGroup(brokerAddress, topic, groupID string, size int) *ConsumerGroup {
	consumers := make([]*Consumer, size)
	for i := 0; i < size; i++ {
		consumers[i] = NewConsumer(brokerAddress, topic, groupID)
	}
	return &ConsumerGroup{
		consumer: consumers,
	}
}

func (g *ConsumerGroup) Consume(ctx context.Context, handler func(key, value []byte) error) {
	for _, c := range g.consumer {
		go c.Consume(ctx, handler)
	}
}

func (g *ConsumerGroup) Close() {
	for _, c := range g.consumer {
		c.Close()
	}
}

package kafka

import (
	"context"
	"sync"
)

type ConsumerGroup struct {
	consumer []*Consumer
	wg       sync.WaitGroup
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
		g.wg.Add(1)
		go func(c *Consumer) {
			defer g.wg.Done()
			c.Consume(ctx, handler)
		}(c)
	}
}

func (g *ConsumerGroup) Close() {
	for _, c := range g.consumer {
		c.Close()
	}
}

func (g *ConsumerGroup) Wait() {
	g.wg.Wait()
}

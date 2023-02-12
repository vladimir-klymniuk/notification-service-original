package producer

import (
	"context"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_publisher_Publish(t *testing.T) {
	config := sarama.NewConfig()
	// set mock producer to expect success
	config.Producer.Return.Successes = true

	producer := mocks.NewAsyncProducer(t, config)
	// again, set mock to expect success
	producer.ExpectInputAndSucceed()

	ts := time.Unix(1, 2)

	p := &publisher{
		producer:  producer,
		timestamp: func() time.Time { return ts },
	}

	message := []byte("hello sarama")

	err := p.Publish(context.Background(), message)
	assert.NoError(t, err)

	select {
	case s := <-producer.Successes():
		assert.Equal(t, sarama.ByteEncoder(message), s.Value)
		assert.Equal(t, ts, s.Timestamp)
	case <-time.After(1 * time.Second):
		assert.Fail(t, "timeout")
	}

	assert.NoError(t, producer.Close())
}

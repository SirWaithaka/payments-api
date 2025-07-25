package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	pkgerrors "github.com/SirWaithaka/payments-api/internal/pkg/errors"
)

// ProducerConfig specific configuration for producers.
type ProducerConfig struct {
	BatchSize    int
	BatchTimeout int // In milliseconds
	Async        bool
	RequiredAcks kafka.RequiredAcks
	// max number of retries when posting a message
	PostRetries int
	// duration to delay before making a retry
	RetryDelay time.Duration
}

// Producer is a Kafka producer structure.
type Producer struct {
	writer      *kafka.Writer
	postRetries int
	retryDelay  time.Duration
}

// NewProducer creates a new producer instance.
func NewProducer(cfg Config, pCfg ProducerConfig) *Producer {
	return &Producer{
		postRetries: pCfg.PostRetries,
		retryDelay:  pCfg.RetryDelay,
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(cfg.Brokers...),
			Balancer:               &kafka.LeastBytes{},
			BatchSize:              pCfg.BatchSize,
			BatchTimeout:           time.Duration(pCfg.BatchTimeout) * time.Millisecond,
			RequiredAcks:           pCfg.RequiredAcks,
			Async:                  pCfg.Async,
			AllowAutoTopicCreation: true,
		},
	}
}

// SendMessage sends a message to the Kafka topic.
func (p *Producer) SendMessage(ctx context.Context, topic string, key, value []byte) error {
	l := zerolog.Ctx(ctx)

	var (
		err          error
		retryCount   = 0
		currentDelay = p.retryDelay
	)

	for retryCount = range p.postRetries {
		l.Debug().Msgf("sending message, retry: %d", retryCount)
		// post message
		err = p.writer.WriteMessages(ctx, kafka.Message{
			Topic: topic,
			Key:   key,
			Value: value,
		})

		// return immediately if err is nil
		if err == nil {
			l.Debug().Msg("message sent")
			return nil
		}

		// check if the error is temporary, so we can retry
		var te pkgerrors.Temporary
		if errors.As(err, &te) && !te.Temporary() || retryCount == p.postRetries {
			// if error is not temporary or retry count
			l.Error().Err(err).Msg("failed to send message")
			break
		}

		// calculate the next retry delay with exponential backoff
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(currentDelay):
			currentDelay = time.Duration(float64(currentDelay) * 1.5)
		}
	}

	return err
}

// Close closes the producer writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}

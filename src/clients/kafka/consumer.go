package kafka

import (
	"bytes"
	"context"
	"errors"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	"github.com/SirWaithaka/payments-api/pkg/logger"
	"github.com/SirWaithaka/payments-api/src/events"
)

// ConsumerConfig specific configuration for consumers.
type ConsumerConfig struct {
	Brokers        []string
	Topic          string
	GroupID        string
	Partition      int
	MinBytes       int
	MaxBytes       int
	CommitInterval int // In milliseconds
	StartOffset    int64
}

// Consumer is a Kafka consumer structure.
type Consumer struct {
	reader  *kafka.Reader
	handler events.Handler
}

// NewConsumer creates a new consumer instance.
func NewConsumer(cCfg ConsumerConfig) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        cCfg.Brokers,
			Topic:          cCfg.Topic,
			GroupID:        cCfg.GroupID,
			MinBytes:       cCfg.MinBytes,
			MaxBytes:       cCfg.MaxBytes,
			CommitInterval: time.Duration(cCfg.CommitInterval) * time.Millisecond,
			StartOffset:    cCfg.StartOffset,
		}),
	}
}

// ReadMessage reads a message from the Kafka topic.
func (c *Consumer) ReadMessage(ctx context.Context) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Msgf("starting consumer - %s on host %s", c.reader.Config().Topic, c.reader.Config().Brokers[0])

loop:
	for {
		// check if context is canceled
		select {
		case <-ctx.Done():
			l.Debug().Msgf("stopping consumer - %s", c.reader.Config().Topic)
			break loop
		default:
		}

		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			l.Error().Err(err).Msg("failed to read message")
			continue
		}
		l.Info().Msgf("message at offset %d", message.Offset)

		// get the event handler
		out, fn := c.handler()
		if fn == nil {
			l.Debug().Msg("return handler func is nil")
			return errors.New("handler func is nil")
		}

		// parse event payload into the out var
		if e := jsoniter.NewDecoder(bytes.NewReader(message.Value)).Decode(out); e != nil {
			l.Error().Err(e).Msg("json error parsing message into variable")
			continue
		}
		l.Debug().Any(logger.LData, out).Msg("message parsed")

		// call the event handler function
		e := fn(ctx)
		if e != nil {
			// any error that occurs in the event handler is just logged
			l.Debug().Err(e).Msg("message handler error")
			continue
		}
	}

	return nil
}

func (c *Consumer) SetHandler(handler events.Handler) {
	c.handler = handler
}

// Close closes the consumer reader.
func (c *Consumer) Close() error {
	return c.reader.Close()
}

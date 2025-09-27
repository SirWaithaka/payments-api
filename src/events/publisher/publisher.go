package publisher

import (
	"context"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"

	pkgevents "github.com/SirWaithaka/payments-api/pkg/events"
	"github.com/SirWaithaka/payments-api/pkg/logger"
	kafkaclient "github.com/SirWaithaka/payments-api/src/clients/kafka"
	"github.com/SirWaithaka/payments-api/src/config"
)

func New(kCfg config.KafkaConfig) Publisher {

	// kafka client config
	brokers := strings.Split(kCfg.Host, ",")
	cfg := kafkaclient.Config{Brokers: brokers}
	// Define producer-specific configuration
	pCfg := kafkaclient.ProducerConfig{
		BatchSize:    100,
		BatchTimeout: 50, // in milliseconds
		Async:        false,
		RequiredAcks: kafka.RequireAll,
		PostRetries:  10,
		RetryDelay:   time.Millisecond * 100,
	}

	producer := kafkaclient.NewProducer(cfg, pCfg)

	return Publisher{producer: producer}

}

type Publisher struct {
	producer *kafkaclient.Producer
}

func (publisher Publisher) Publish(ctx context.Context, event pkgevents.EventType) error {
	l := zerolog.Ctx(ctx)
	l.Debug().Any(logger.LData, event).Msg("event to publish")

	edata, err := jsoniter.Marshal(event)
	if err != nil {
		l.Error().Err(err).Msg("failed to marshal event")
		return err
	}

	err = publisher.producer.SendMessage(ctx, event.Name(), []byte(event.Name()), edata)
	if err != nil {
		l.Error().Err(err).Msg("failed to publish event")
		return err
	}
	l.Info().Msg("message sent")

	return nil
}

func (publisher Publisher) Close() error {
	return publisher.producer.Close()
}

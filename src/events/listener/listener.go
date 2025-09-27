package listener

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"

	"github.com/SirWaithaka/payments-api/pkg/events/subjects"
	"github.com/SirWaithaka/payments-api/pkg/logger"
	kafkaclient "github.com/SirWaithaka/payments-api/src/clients/kafka"
	"github.com/SirWaithaka/payments-api/src/config"
	dipkg "github.com/SirWaithaka/payments-api/src/di"
	"github.com/SirWaithaka/payments-api/src/events"
	"github.com/SirWaithaka/payments-api/src/events/handlers"
)

func New(c context.Context, cfg config.KafkaConfig, di *dipkg.DI) *Listener {
	// create instance of consumers
	consumers := make(map[string][]*kafkaclient.Consumer)

	return &Listener{ctx: c, kafkaCfg: cfg, di: di, consumers: consumers}
}

type Listener struct {
	di        *dipkg.DI
	ctx       context.Context
	kafkaCfg  config.KafkaConfig
	consumers map[string][]*kafkaclient.Consumer

	// a wait group for running consumer goroutines
	waitGroup *errgroup.Group
}

func (listener *Listener) newConsumer(topic string) *kafkaclient.Consumer {
	// kafka client config
	brokers := strings.Split(listener.kafkaCfg.Host, ",")
	// Define consumer-specific configuration
	cCfg := kafkaclient.ConsumerConfig{
		Topic:   topic,
		Brokers: brokers,
		// groupId in the format of <topicName-group>
		GroupID:        fmt.Sprintf("%s-group", topic), //
		Partition:      0,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: 1000, // 1 second
		StartOffset:    kafka.FirstOffset,
	}

	// Initialize the consumer
	return kafkaclient.NewConsumer(cCfg)
}

// RegisterHandler adds a handler to a topic/event name
func (listener *Listener) RegisterHandler(name string, handler events.Handler) {
	// get all consumers for a particular event name/topic
	consumers := listener.consumers[name]
	// create new consumer with given handler
	consumer := listener.newConsumer(name)
	consumer.SetHandler(handler)
	// add new consumer to the list of consumers
	consumers = append(consumers, consumer)
	// update registered consumers
	listener.consumers[name] = consumers
}

func (listener *Listener) Listen() error {
	l := zerolog.Ctx(listener.ctx)
	l.Info().Msg("starting listener")

	handler := handlers.NewHandler(listener.di.Webhook)

	// register event handlers
	listener.RegisterHandler(subjects.WebhookReceived, handler.WebhookReceived)

	g, ctx := errgroup.WithContext(listener.ctx)
	listener.waitGroup = g

	// loop through consumers and start them
	for _, consumers := range listener.consumers {
		for _, consumer := range consumers {
			// run each consumer in a goroutine
			g.Go(func() error {
				// recover from any panics
				defer func() {
					if r := recover(); r != nil {
						l.WithLevel(zerolog.FatalLevel).Str(logger.LData, string(debug.Stack())).Msg("recovered from panic")
					}
				}()

				// read messages from kafka
				return consumer.ReadMessage(ctx)
			})
		}
	}

	return nil
}

func (listener *Listener) Close() error {
	l := zerolog.Ctx(listener.ctx)
	defer l.Info().Msg("listener closed")

	// wait for the done event from context
	<-listener.ctx.Done()
	l.Info().Msg("stopping listener")

	var errs []error

	// block and wait for all consumer goroutines to finish
	err := listener.waitGroup.Wait()
	if err != nil {
		errs = append(errs, err)
	}

	// fetch all consumers and close each
	for _, consumers := range listener.consumers {
		for _, consumer := range consumers {
			if err = consumer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

package events

import (
	"context"

	"github.com/SirWaithaka/payments-api/pkg/events"
)

// Publisher defines a method to publish an event
type Publisher interface {
	Publish(ctx context.Context, event events.EventType) error
}

// Subscriber defines the behavior for subscribing an event
type Subscriber interface {
	Subscribe(ctx context.Context, event events.EventType) error
}

// Handler is a decorator function that returns 2 values
// 1. a variable that satisfies events.EventMessage interface
// 2. a handler function that acts upon the passed event
type Handler func() (events.EventMessage, func(context.Context) error)

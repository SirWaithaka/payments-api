package events

import (
	"time"
)

type EventMessage interface {
	Message() any
}

type EventType interface {
	ID() string
	Name() string
	EventMessage
}

type eventType struct {
	id   string
	name string
}

func (e eventType) ID() string {
	return e.id
}

func (e eventType) Name() string {
	return e.name
}

func (e *eventType) SetID(id string) {
	e.id = id
}

func (e *eventType) SetName(name string) {
	e.name = name
}

type Event[T any] struct {
	eventType
	RequestID   string    `json:"request_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	PublishTime time.Time `json:"publish_time,omitempty"`
	Payload     T         `json:"payload"`
}

func (e Event[T]) Message() any {
	return e.Payload
}

func (e *Event[T]) SetRequestID(id string) {
	e.RequestID = id
}

func (e *Event[T]) SetPublishTime(t time.Time) {
	e.PublishTime = t
}

func NewEvent[T any](name string, payload T) *Event[T] {
	evt := Event[T]{
		CreatedAt: time.Now().UTC(),
		Payload:   payload,
	}

	evt.SetName(name)

	return &evt
}

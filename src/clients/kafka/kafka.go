package kafka

import (
	"context"
)

// Config common configuration used by both producers and consumers.
type Config struct {
	Brokers []string
	Topic   string
}

func New(ctx context.Context, host string) {

}

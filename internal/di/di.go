package di

import (
	"github.com/SirWaithaka/payments-api/internal/config"
)

type DI struct {
	Cfg *config.Config
}

func New(cfg config.Config) *DI {
	return &DI{&cfg}
}

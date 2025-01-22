package worker

import (
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/wa-worker/core"
	"gopkg.in/validator.v2"
)

type Worker struct {
	Config
}

type Config struct {
	Service core.Service `validate:"nonnil"`
}

func New(cfg Config) (*Worker, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &Worker{Config: cfg}, nil
}

// Run starts the worker and block until it's done.
func (w *Worker) Run() error {
	// TODO
	return nil
}

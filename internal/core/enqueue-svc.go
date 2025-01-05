package core

import (
	"context"
	"fmt"

	"gopkg.in/validator.v2"
)

type EnqueueService interface {
	Enqueue(ctx context.Context, req SubmitRequest) error
}

type EnqueueServiceConfig struct {
	Queue Queue `validate:"nonnil"`
}

func NewEnqueueService(cfg EnqueueServiceConfig) (EnqueueService, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &enqueueService{
		EnqueueServiceConfig: cfg,
	}, nil
}

type enqueueService struct {
	EnqueueServiceConfig
}

func (s *enqueueService) Enqueue(ctx context.Context, req SubmitRequest) error {
	return s.Queue.Enqueue(ctx, req)
}

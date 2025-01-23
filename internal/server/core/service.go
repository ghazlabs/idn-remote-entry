package core

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type Service interface {
	HandleRequest(ctx context.Context, req core.SubmitRequest) error
}

type ServiceConfig struct {
	Queue Queue `validate:"nonnil"`
}

func NewService(cfg ServiceConfig) (Service, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &service{
		ServiceConfig: cfg,
	}, nil
}

type service struct {
	ServiceConfig
}

func (s *service) HandleRequest(ctx context.Context, req core.SubmitRequest) error {
	// validate request
	err := validator.Validate(req)
	if err != nil {
		return core.NewBadRequestError(fmt.Sprintf("invalid request: %v", err))
	}

	// put request to queue
	err = s.Queue.Put(ctx, req)
	if err != nil {
		return core.NewInternalError(err)
	}

	return nil
}

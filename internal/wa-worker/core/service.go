package core

import (
	"context"
	"fmt"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type Service interface {
	Handle(ctx context.Context, n shcore.WaNotification) error
}

type ServiceConfig struct {
	Publisher Publisher `validate:"nonnil"`
}

func NewService(cfg ServiceConfig) (Service, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &service{
		ServiceConfig: cfg,
	}, nil
}

type service struct {
	ServiceConfig
}

func (s *service) Handle(ctx context.Context, n shcore.WaNotification) error {
	err := s.Publisher.Publish(ctx, n)
	if err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}
	return nil
}

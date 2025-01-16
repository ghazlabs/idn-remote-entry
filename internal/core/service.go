package core

import (
	"context"
	"fmt"

	"gopkg.in/validator.v2"
)

type Service interface {
	Handle(ctx context.Context, req SubmitRequest) error
}

type ServiceConfig struct {
	Storage         Storage         `validate:"nonnil"`
	VacancyResolver VacancyResolver `validate:"nonnil"`
	Notifier        Notifier        `validate:"nonnil"`
}

func NewService(cfg ServiceConfig) (Service, error) {
	// validate config
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

func (s *service) Handle(ctx context.Context, req SubmitRequest) error {
	// validate request ensure it is valid request
	err := req.Validate()
	if err != nil {
		return NewBadRequestError(err.Error())
	}

	// handle request
	switch req.SubmissionType {
	case SubmitTypeManual:
	case SubmitTypeURL:
		// resolve apply url to get vacancy details
		v, err := s.VacancyResolver.Resolve(ctx, req.Vacancy.ApplyURL)
		if err != nil {
			// the error will be determined by resolver implementation
			return err
		}

		// update vacancy with resolved details
		req.Vacancy = *v
	}

	// save vacancy
	rec, err := s.Storage.Save(ctx, req.Vacancy)
	if err != nil {
		return NewInternalError(fmt.Errorf("failed to save vacancy: %w", err))
	}

	// notify to whatsapp
	err = s.Notifier.Notify(ctx, *rec)
	if err != nil {
		return NewInternalError(fmt.Errorf("failed to notify to whatsapp: %w", err))
	}

	return nil
}

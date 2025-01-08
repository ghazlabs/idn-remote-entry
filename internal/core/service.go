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
		err = s.Storage.Save(ctx, req.Vacancy)
		if err != nil {
			return NewInternalError(err)
		}
	case SubmitTypeURL:
		// resolve apply url to get vacancy details
		vacancy, err := s.VacancyResolver.Resolve(ctx, req.Vacancy.ApplyURL)
		if err != nil {
			// the error will be determined by resolver implementation
			return err
		}

		// then save the job details
		err = s.Storage.Save(ctx, *vacancy)
		if err != nil {
			return NewInternalError(err)
		}
	}

	return nil
}

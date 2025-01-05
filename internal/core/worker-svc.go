package core

import (
	"context"
	"fmt"

	"gopkg.in/validator.v2"
)

type WorkerService interface {
	Execute(ctx context.Context) error
}

type WorkerServiceConfig struct {
	Storage         Storage         `validate:"nonnil"`
	VacancyResolver VacancyResolver `validate:"nonnil"`
	Queue           Queue           `validate:"nonnil"`
}

func NewWorkerService(cfg WorkerServiceConfig) (WorkerService, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &workerService{
		WorkerServiceConfig: cfg,
	}, nil
}

type workerService struct {
	WorkerServiceConfig
}

func (s *workerService) Execute(ctx context.Context) error {
	// fetch submit request from queue
	req, err := s.Queue.Dequeue(ctx)
	if err != nil {
		return fmt.Errorf("unable to dequeue due: %w", err)
	}

	// if there is no request in the queue just return
	if req == nil {
		return nil
	}

	// validate request ensure it is valid request
	err = req.Validate()
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

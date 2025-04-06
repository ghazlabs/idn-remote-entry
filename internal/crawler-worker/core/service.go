package core

import (
	"context"
	"fmt"
	"log"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type Service interface {
	Run(ctx context.Context) error
}

type ServiceConfig struct {
	Crawler        Crawler          `validate:"nonzero"`
	VacancyStorage VacanciesStorage `validate:"nonzero"`
	Server         Server           `validate:"nonzero"`
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

func (s *service) Run(ctx context.Context) error {
	vacancies, err := s.Crawler.Crawl(ctx)
	if err != nil {
		return fmt.Errorf("failed to crawl: %w", err)
	}

	filteredVacancies := make([]core.Vacancy, 0)
	for _, v := range vacancies {
		if s.VacancyStorage.IsVacancyExists(ctx, v) {
			continue
		}
		filteredVacancies = append(filteredVacancies, v)
	}

	for _, v := range filteredVacancies {
		err = s.Server.SubmitVacancy(ctx, v)
		if err != nil {
			log.Printf("failed to submit vacancy: %s, error: %v", v.ToJSON(), err)
			// skip error
			continue
		}
	}

	return nil
}

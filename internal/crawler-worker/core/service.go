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
	Crawler         Crawler          `validate:"nonzero"`
	VacancyStorage  VacanciesStorage `validate:"nonzero"`
	ContentChecker  ContentChecker   `validate:"nonzero"`
	ApprovalStorage ApprovalStorage  `validate:"nonnil"`
	Server          Server           `validate:"nonzero"`

	EnabledApplicableChecker bool
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

	log.Printf("found %d vacancies\n", len(vacancies))

	// It more easy to get all url vacancies from storage and filter out already existing vacancies
	// than to check each vacancy if it already exists in storage
	allURLVacancies, err := s.VacancyStorage.GetAllURLVacancies(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all urls vacancies: %w", err)
	}

	log.Printf("total vacancies in storage: %d", len(allURLVacancies))

	filterVacancies := make([]core.Vacancy, 0)
	for _, v := range vacancies {
		// check if vacancy already exists
		if _, ok := allURLVacancies[v.ApplyURL]; ok {
			// skip if vacancy already exists
			continue
		}

		// check if vacancy already requested (avoid duplicate request)
		isAlreadyRequested, err := s.ApprovalStorage.IsVacancyAlreadyRequested(ctx, v.ApplyURL)
		if err != nil {
			log.Printf("failed to check if vacancy is already requested: %s, error: %v", v.ToJSON(), err)
			// skip error
			continue
		}
		
		if isAlreadyRequested {
			continue
		}

		// check if vacancy is applicable for Indonesian
		if s.EnabledApplicableChecker {
			isApplicable, err := s.ContentChecker.IsApplicable(ctx, v)
			if err != nil {
				log.Printf("failed to check if vacancy is applicable for indonesian: %s, error: %v", v.ToJSON(), err)
				// skip error
				continue
			}

			if !isApplicable {
				continue
			}
		}

		// add vacancy to filter vacancies
		filterVacancies = append(filterVacancies, v)
	}

	if len(filterVacancies) == 0 {
		log.Println("no vacancies to submit")
		return nil
	}

	err = s.Server.SubmitBulkVacancies(ctx, filterVacancies)
	if err != nil {
		return fmt.Errorf("failed to submit bulk vacancies: %w", err)
	}

	return nil
}

package core

import (
	"context"
	"fmt"
	"log"

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

	// Print all vacancies found during crawling
	for i, v := range vacancies {
		log.Printf("Vacancy #%d: Title=%s, Company=%s, URL=%s",
			i+1, v.JobTitle, v.CompanyName, v.ApplyURL)
	}

	log.Printf("found %d vacancies\n", len(vacancies))

	// It more easy to get all url vacancies from storage and filter out already existing vacancies
	// than to check each vacancy if it already exists in storage
	allURLVacancies, err := s.VacancyStorage.GetAllURLVacancies(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all urls vacancies: %w", err)
	}

	log.Printf("total vacancies in storage: %d", len(allURLVacancies))

	// Track submission statistics
	skippedAlreadyExists := 0
	skippedAlreadyRequested := 0
	skippedNotApplicable := 0
	skippedSubmitError := 0
	submittedCount := 0

	for _, v := range vacancies {
		// check if vacancy already exists
		if _, ok := allURLVacancies[v.ApplyURL]; ok {
			// skip if vacancy already exists
			skippedAlreadyExists++
			continue
		}

		// check if vacancy already requested (avoid duplicate request)
		isRequested, err := s.ApprovalStorage.IsVacancyAlreadyRequested(ctx, v.ApplyURL)
		if err != nil {
			log.Printf("failed to check if vacancy is already requested: %s, error: %v", v.ToJSON(), err)
			// skip error
			continue
		}
		if isRequested {
			skippedAlreadyRequested++
			continue
		}

		// check if vacancy is applicable for Indonesian
		if s.EnabledApplicableChecker {
			isApplicable, err := s.ContentChecker.IsApplicableForIndonesian(ctx, v)
			if err != nil {
				log.Printf("failed to check if vacancy is applicable for indonesian: %s, error: %v", v.ToJSON(), err)
				// skip error
				continue
			}

			if !isApplicable {
				skippedNotApplicable++
				continue
			}
		}

		err = s.Server.SubmitURLVacancy(ctx, v.ApplyURL)
		if err != nil {
			log.Printf("failed to submit vacancy: %s, error: %v", v.ToJSON(), err)
			skippedSubmitError++
			// skip error
			continue
		}
		submittedCount++
	}

	log.Printf("submission summary: total=%d, submitted=%d, skipped_already_exists=%d, skipped_already_requested=%d, skipped_not_applicable=%d, skipped_submit_error=%d",
		len(vacancies), submittedCount, skippedAlreadyExists, skippedAlreadyRequested, skippedNotApplicable, skippedSubmitError)

	return nil
}

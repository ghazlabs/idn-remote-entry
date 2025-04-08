package core

import (
	"context"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type Crawler interface {
	Crawl(ctx context.Context) ([]core.Vacancy, error)
}

type VacanciesStorage interface {
	GetAllURLVacancies(ctx context.Context) (map[string]bool, error)
}

type ContentChecker interface {
	IsApplicable(ctx context.Context, v core.Vacancy) (bool, error)
}

type Server interface {
	SubmitURLVacancy(ctx context.Context, applyURL string) error
}

type ApprovalStorage interface {
	IsVacancyAlreadyRequested(ctx context.Context, applyURL string) (bool, error)
}

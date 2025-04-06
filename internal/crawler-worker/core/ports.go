package core

import (
	"context"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type Crawler interface {
	Crawl(ctx context.Context) ([]core.Vacancy, error)
}

type VacanciesStorage interface {
	IsVacancyExists(ctx context.Context, vacancy core.Vacancy) bool
}

type Server interface {
	SubmitVacancy(ctx context.Context, vacancy core.Vacancy) error
}

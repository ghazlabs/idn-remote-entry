package core

import (
	"context"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type Storage interface {
	Save(ctx context.Context, v shcore.Vacancy) (*shcore.VacancyRecord, error)
	LookupCompanyLocation(ctx context.Context, companyName string) (string, error)
}

type VacancyResolver interface {
	Resolve(ctx context.Context, url string) (*shcore.Vacancy, error)
}

type Notifier interface {
	Notify(ctx context.Context, v shcore.VacancyRecord) error
}

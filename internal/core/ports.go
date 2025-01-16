package core

import "context"

type Storage interface {
	Save(ctx context.Context, v Vacancy) (*VacancyRecord, error)
}

type VacancyResolver interface {
	Resolve(ctx context.Context, url string) (*Vacancy, error)
}

type Notifier interface {
	Notify(ctx context.Context, v VacancyRecord) error
}

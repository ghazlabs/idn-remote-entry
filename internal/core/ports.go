package core

import "context"

type Storage interface {
	Save(ctx context.Context, v Vacancy) error
}

type VacancyResolver interface {
	Resolve(ctx context.Context, url string) (*Vacancy, error)
}

type Queue interface {
	Enqueue(ctx context.Context, req SubmitRequest) error
	Dequeue(ctx context.Context) (*SubmitRequest, error)
}

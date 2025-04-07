package crawler

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type VacancyCrawler struct {
	VacancyResolverConfig
}

type Crawler interface {
	Crawl(ctx context.Context) ([]core.Vacancy, error)
}

type CrawlRegistry struct {
	Name    string
	Crawler Crawler
}

type VacancyResolverConfig struct {
	CrawlerRegistries []CrawlRegistry
}

func NewVacancyCrawler(cfg VacancyResolverConfig) (*VacancyCrawler, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &VacancyCrawler{
		VacancyResolverConfig: cfg,
	}, nil
}

func (r *VacancyCrawler) Crawl(ctx context.Context) ([]core.Vacancy, error) {
	vacancies := make([]core.Vacancy, 0)
	for _, reg := range r.CrawlerRegistries {
		vac, err := reg.Crawler.Crawl(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to crawl from %s due to: %w", reg.Name, err)
		}

		for _, v := range vac {
			err = v.Validate()
			// skip invalid vacancy
			if err != nil {
				continue
			}

			if isEligible := isEligibleToSave(v); !isEligible {
				continue
			}

			vacancies = append(vacancies, v)
		}
	}

	return vacancies, nil
}

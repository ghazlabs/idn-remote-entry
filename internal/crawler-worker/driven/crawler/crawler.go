package crawler

import (
	"context"
	"fmt"
	"log"

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
	allVacancies := make([]core.Vacancy, 0)

	for _, reg := range r.CrawlerRegistries {
		log.Printf("Starting to crawl from %s", reg.Name)

		vac, err := reg.Crawler.Crawl(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to crawl from %s due to: %w", reg.Name, err)
		}

		log.Printf("Found %d vacancies from %s", len(vac), reg.Name)

		// Track vacancies from this crawler only
		crawlerVacancies := make([]core.Vacancy, 0)

		for _, v := range vac {
			err = v.Validate()
			// skip invalid vacancy
			if err != nil {
				log.Printf("Skipping invalid vacancy from %s: %v", reg.Name, err)
				continue
			}

			if !isEligibleToSave(v) {
				continue
			}

			v.ShortDescription = removeUnwantedText(v.ShortDescription)

			crawlerVacancies = append(crawlerVacancies, v)
		}

		log.Printf("Added %d valid vacancies from %s", len(crawlerVacancies), reg.Name)
		allVacancies = append(allVacancies, crawlerVacancies...)
	}

	log.Printf("Total valid vacancies from all crawlers: %d", len(allVacancies))
	return allVacancies, nil
}

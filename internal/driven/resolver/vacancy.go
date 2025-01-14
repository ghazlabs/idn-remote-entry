package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"gopkg.in/validator.v2"
)

type VacancyResolver struct {
	VacancyResolverConfig
}

type Parser interface {
	Parse(ctx context.Context, url string) (*core.Vacancy, error)
}

type ParserRegistry struct {
	ApexDomains []string
	Parser      Parser
}

type VacancyResolverConfig struct {
	DefaultParser    Parser `validate:"nonnil"`
	ParserRegistries []ParserRegistry
}

func NewVacancyResolver(cfg VacancyResolverConfig) (*VacancyResolver, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &VacancyResolver{
		VacancyResolverConfig: cfg,
	}, nil
}

func (r *VacancyResolver) Resolve(ctx context.Context, url string) (*core.Vacancy, error) {
	var vac *core.Vacancy
	var err error
	for _, reg := range r.ParserRegistries {
		for _, apex := range reg.ApexDomains {
			if strings.Contains(url, apex) {
				vac, err = reg.Parser.Parse(ctx, url)
				if err != nil {
					return nil, fmt.Errorf("failed to parse the vacancy: %w", err)
				}

				goto parserFound
			}
		}
	}

	// if no parser found, use the default parser
	vac, err = r.DefaultParser.Parse(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the vacancy: %w", err)
	}

parserFound:
	return vac, nil
}

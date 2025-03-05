package resolver

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type VacancyResolver struct {
	VacancyResolverConfig
}

type Parser interface {
	Parse(ctx context.Context, url string) (*core.Vacancy, error)
}

type HQLocator interface {
	Locate(ctx context.Context, companyName string) (string, error)
}

type ParserRegistry struct {
	ApexDomains []string
	Parser      Parser
}

type VacancyResolverConfig struct {
	DefaultParser    Parser `validate:"nonnil"`
	ParserRegistries []ParserRegistry
	HQLocator        HQLocator `validate:"nonnil"`
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
				// parse the vacancy from URL
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
	err = vac.Validate()
	if err != nil {
		return nil, core.NewBadRequestError(err.Error())
	}

	// if the company location is not found (which indicated by "Global Remote")
	// locate the company's headquarters
	if strings.Contains(strings.ToLower(vac.CompanyLocation), "remote") {
		hqLoc, _ := r.HQLocator.Locate(ctx, vac.CompanyName)
		if len(hqLoc) > 0 {
			// update if the company location is found
			vac.CompanyLocation = hqLoc
		}
	}

	return vac, nil
}

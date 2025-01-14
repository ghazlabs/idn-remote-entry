package resolver_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	vacResolver, err := resolver.NewVacancyResolver(resolver.VacancyResolverConfig{
		HttpClient:    resty.New(),
		OpenaAiClient: openai.NewClient(),
	})
	require.NoError(t, err)

	testCases := []struct {
		Name           string
		VacancyURL     string
		ExpVacancyName string
		ExpCompanyName string
	}{
		// {
		// 	Name:           "Invertase URL",
		// 	VacancyURL:     "https://job-boards.eu.greenhouse.io/invertase/jobs/4492621101",
		// 	ExpVacancyName: "Staff Software Engineer - Cloud Platforms",
		// 	ExpCompanyName: "Invertase",
		// },
		// {
		// 	Name:           "Matter Labs URL",
		// 	VacancyURL:     "https://jobs.ashbyhq.com/matter-labs/f6054a2f-ea5d-42ee-a243-8c3fa95018ef",
		// 	ExpVacancyName: "Senior Protocol Engineer",
		// 	ExpCompanyName: "Matter Labs",
		// },
		// {
		// 	Name:           "Automattic URL",
		// 	VacancyURL:     "https://boards.greenhouse.io/automatticcareers/jobs/6510955",
		// 	ExpVacancyName: "Code Wrangler - Support Tooling",
		// 	ExpCompanyName: "Automattic",
		// },
		// {
		// 	Name:           "Remote.com URL",
		// 	VacancyURL:     "https://job-boards.greenhouse.io/remotecom/jobs/6322023003",
		// 	ExpVacancyName: "Lifecycle Specialist: Contracts Management",
		// 	ExpCompanyName: "Remote",
		// },
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			vac, err := vacResolver.Resolve(context.Background(), testCase.VacancyURL)
			require.NoError(t, err)

			assert.Equal(t, strings.ToLower(testCase.ExpVacancyName), strings.ToLower(vac.JobTitle))
			assert.Equal(t, strings.ToLower(testCase.ExpCompanyName), strings.ToLower(vac.CompanyName))
			assert.NotEmpty(t, vac.ShortDescription)
			assert.NotEmpty(t, vac.RelevantTags)
		})
	}

}

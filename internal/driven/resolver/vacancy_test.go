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
		{
			Name:           "Greenhouse URL",
			VacancyURL:     "https://job-boards.eu.greenhouse.io/invertase/jobs/4492528101",
			ExpVacancyName: "Developer Advocate - Dart & Flutter Products",
			ExpCompanyName: "Invertase",
		},
		{
			Name:           "Greenhouse URL 2",
			VacancyURL:     "https://job-boards.eu.greenhouse.io/invertase/jobs/4492621101",
			ExpVacancyName: "Staff Software Engineer - Cloud Platforms",
			ExpCompanyName: "Invertase",
		},
		{
			Name:           "AshbyHQ URL",
			VacancyURL:     "https://jobs.ashbyhq.com/matter-labs/f6054a2f-ea5d-42ee-a243-8c3fa95018ef",
			ExpVacancyName: "Senior Protocol Engineer",
			ExpCompanyName: "Matter Labs",
		},
		{
			Name:           "YCombinator URL",
			VacancyURL:     "https://www.ycombinator.com/companies/akiflow/jobs/yMcXc5g-senior-mobile-developer-flutter",
			ExpVacancyName: "Senior Mobile Developer - Flutter",
			ExpCompanyName: "Akiflow",
		},
		{
			Name:           "Micro1 URL",
			VacancyURL:     "https://jobs.micro1.ai/post/ee6e8b24-ae08-472f-863b-aabcb1b25cef",
			ExpVacancyName: "AI Engineer",
			ExpCompanyName: "micro1",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			vac, err := vacResolver.Resolve(context.Background(), testCase.VacancyURL)
			require.NoError(t, err)

			assert.Equal(t, strings.ToLower(testCase.ExpVacancyName), strings.ToLower(vac.JobTitle))
			assert.Equal(t, strings.ToLower(testCase.ExpCompanyName), strings.ToLower(vac.CompanyName))
			assert.NotEmpty(t, vac.ShortDescription)
			assert.NotEmpty(t, vac.RelevantTags)
			assert.NotEmpty(t, vac.CompanyLocation)
		})
	}

}

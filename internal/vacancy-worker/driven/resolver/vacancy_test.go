package resolver_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/hqloc"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/parser"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	httpClient := resty.New()
	openAiClient := openai.NewClient(option.WithAPIKey(env.GetString(testutil.EnvKeyTestOpenAiKey)))
	textParser, err := parser.NewGreenhouseParser(parser.GreenhouseParserConfig{
		HttpClient:    httpClient,
		OpenaAiClient: openAiClient,
	})
	require.NoError(t, err)

	ocrParser, err := parser.NewOCRParser(parser.OCRParserConfig{
		OpenaAiClient: openAiClient,
	})
	require.NoError(t, err)

	locator, err := hqloc.NewLocator(hqloc.LocatorConfig{
		OpenaAiClient: openai.NewClient(option.WithAPIKey(env.GetString(testutil.EnvKeyTestOpenAiKey))),
	})
	require.NoError(t, err)

	vacResolver, err := resolver.NewVacancyResolver(resolver.VacancyResolverConfig{
		ParserRegistries: []resolver.ParserRegistry{
			{
				ApexDomains: []string{"greenhouse.io"},
				Parser:      textParser,
			},
		},
		DefaultParser: ocrParser,
		HQLocator:     locator,
	})
	require.NoError(t, err)

	testCases := []struct {
		Name           string
		VacancyURL     string
		ExpVacancyName string
		ExpCompanyName string
	}{
		{
			Name:           "Invertase URL",
			VacancyURL:     "https://job-boards.eu.greenhouse.io/invertase/jobs/4492621101",
			ExpVacancyName: "Staff Software Engineer - Cloud Platforms",
			ExpCompanyName: "Invertase",
		},
		{
			Name:           "Matter Labs URL",
			VacancyURL:     "https://jobs.ashbyhq.com/matter-labs/f6054a2f-ea5d-42ee-a243-8c3fa95018ef",
			ExpVacancyName: "Senior Protocol Engineer",
			ExpCompanyName: "Matter Labs",
		},
		{
			Name:           "Automattic URL",
			VacancyURL:     "https://boards.greenhouse.io/automatticcareers/jobs/6510955",
			ExpVacancyName: "Code Wrangler - Support Tooling",
			ExpCompanyName: "Automattic",
		},
		{
			Name:           "Remote.com URL",
			VacancyURL:     "https://job-boards.greenhouse.io/remotecom/jobs/6322023003",
			ExpVacancyName: "Lifecycle Specialist: Contracts Management - APAC",
			ExpCompanyName: "Remote",
		},
		{
			Name:           "Makro Pro URL",
			VacancyURL:     "https://apply.workable.com/joinmakropro/j/A182E331FE/",
			ExpVacancyName: "Backend Engineer, Digital Venture - Fully REMOTE",
			ExpCompanyName: "Makro Pro",
		},
		{
			Name:           "Dynatrace URL",
			VacancyURL:     "https://careers.dynatrace.com/jobs/1243381900/",
			ExpVacancyName: "Sr Customer Success Engineer",
			ExpCompanyName: "Dynatrace",
		},
		{
			Name:           "Goodnotes URL",
			VacancyURL:     "https://job-boards.greenhouse.io/goodnotes/jobs/5158740004",
			ExpVacancyName: "Lead iOS Engineer (Indonesia Remote)",
			ExpCompanyName: "Goodnotes",
		},
		{
			Name:           "Influx URL",
			VacancyURL:     "https://influx.com/careers/jobs/3-customer-service-representative",
			ExpVacancyName: "Customer Service Representative",
			ExpCompanyName: "Influx",
		},
		{
			Name:           "Fingerprint URL",
			VacancyURL:     "https://fingerprint.com/careers/jobs/apply/?gh_jid=5377202004",
			ExpVacancyName: "Sr. Android Engineer",
			ExpCompanyName: "Fingerprint",
		},
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
			assert.NotEmpty(t, vac.ApplyURL)
			assert.NotEmpty(t, vac.CompanyLocation)
		})
	}
}

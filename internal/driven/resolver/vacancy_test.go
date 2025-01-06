package resolver_test

import (
	"context"
	"log"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	vacResolver, err := resolver.NewVacancyResolver(resolver.VacancyResolverConfig{
		HTTPClient:    resty.New(),
		OpenaAiClient: openai.NewClient(),
	})
	require.NoError(t, err)

	vacURL := "https://job-boards.eu.greenhouse.io/invertase/jobs/4492528101"
	//vacURL := "https://jobs.ashbyhq.com/matter-labs/f6054a2f-ea5d-42ee-a243-8c3fa95018ef"
	//vacURL := "https://www.ycombinator.com/companies/akiflow/jobs/yMcXc5g-senior-mobile-developer-flutter"
	vac, err := vacResolver.Resolve(context.Background(), vacURL)
	require.NoError(t, err)

	log.Printf("Vacancy: %+v", vac)

	require.NotEmpty(t, vac.JobTitle)
	require.Equal(t, "Invertase", vac.CompanyName)
	require.NotEmpty(t, vac.ShortDescription)
	require.NotEmpty(t, vac.RelevantTags)
	// require.NotEmpty(t, vac.CompanyLocation)
}

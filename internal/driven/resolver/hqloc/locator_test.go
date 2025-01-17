package hqloc_test

import (
	"context"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver/hqloc"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/require"
)

func TestLocate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	locator, err := hqloc.NewLocator(hqloc.LocatorConfig{
		HttpClient:    resty.New(),
		DatabaseID:    env.GetString(testutil.EnvKeyNotionDatabaseID),
		NotionToken:   env.GetString(testutil.EnvKeyNotionToken),
		OpenaAiClient: openai.NewClient(option.WithAPIKey(env.GetString(testutil.EnvKeyTestOpenAiKey))),
	})
	require.NoError(t, err)

	testCases := []struct {
		CompanyName        string
		ExpCompanyLocation string
	}{
		{
			CompanyName:        "Automattic",
			ExpCompanyLocation: "San Francisco, United States",
		},
		{
			CompanyName:        "Haraj",
			ExpCompanyLocation: "Riyadh, Saudi Arabia",
		},
		{
			CompanyName:        "Fingerprint",
			ExpCompanyLocation: "Chicago, United States",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.CompanyName, func(t *testing.T) {
			loc, err := locator.Locate(context.Background(), testCase.CompanyName)
			require.NoError(t, err)
			require.Equal(t, testCase.ExpCompanyLocation, loc)
		})
	}
}

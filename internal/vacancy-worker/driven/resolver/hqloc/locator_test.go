package hqloc_test

import (
	"context"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/hqloc"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/require"
)

type mockStorage struct {
	companyLocation string
}

func (m *mockStorage) LookupCompanyLocation(ctx context.Context, companyName string) (string, error) {
	return m.companyLocation, nil
}

func (m *mockStorage) Save(ctx context.Context, v core.Vacancy) (*core.VacancyRecord, error) {
	return nil, nil
}

func TestLocate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	mockStorage := &mockStorage{}
	locator, err := hqloc.NewLocator(hqloc.LocatorConfig{
		Storage:       mockStorage,
		ModelLLM:      openai.ChatModelGPT4o2024_08_06,
		OpenaAiClient: openai.NewClient(option.WithAPIKey(env.GetString(testutil.EnvKeyTestOpenAiKey))),
	})
	require.NoError(t, err)

	testCases := []struct {
		CompanyName           string
		CompanyLocFromStorage string
		ExpCompanyLocation    string
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
			CompanyName:           "Fingerprint",
			CompanyLocFromStorage: "Chicago, United States",
			ExpCompanyLocation:    "Chicago, United States",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.CompanyName, func(t *testing.T) {
			if testCase.CompanyLocFromStorage != "" {
				mockStorage.companyLocation = testCase.CompanyLocFromStorage
			}

			loc, err := locator.Locate(context.Background(), testCase.CompanyName)
			require.NoError(t, err)
			require.Equal(t, testCase.ExpCompanyLocation, loc)
		})
	}
}

func TestLocateWithOllama(t *testing.T) {
	ollamaBaseUrl := env.GetString(testutil.EnvKeyOllamaBaseUrl)
	if ollamaBaseUrl == "" {
		t.Skip("skip test because Ollama base URL is not set")
	}

	testCases := []struct {
		Name                  string
		CompanyName           string
		CompanyLocFromStorage string
		ExpCompanyLocation    string
	}{
		{
			Name:               "Return from Ollama",
			CompanyName:        "Haraj",
			ExpCompanyLocation: "Riyadh",
		},
		{
			Name:                  "Return from storage",
			CompanyName:           "Fingerprint",
			CompanyLocFromStorage: "Chicago, United States",
			ExpCompanyLocation:    "Chicago, United States",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			mockStorage := &mockStorage{
				companyLocation: testCase.CompanyLocFromStorage,
			}
			locator, err := hqloc.NewLocator(hqloc.LocatorConfig{
				Storage:       mockStorage,
				ModelLLM:      env.GetString(testutil.EnvKeyOllamaModel),
				OpenaAiClient: openai.NewClient(option.WithBaseURL(ollamaBaseUrl)),
			})

			loc, err := locator.Locate(context.Background(), testCase.CompanyName)
			require.NoError(t, err)
			require.Contains(t, loc, testCase.ExpCompanyLocation)
		})
	}
}

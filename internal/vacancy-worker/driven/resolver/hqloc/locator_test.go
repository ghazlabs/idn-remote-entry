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
	type testCase struct {
		Name                  string
		CompanyName           string
		CompanyLocFromStorage string
		ExpCompanyLocation    string
	}
	testCases := []testCase{
		{
			Name:                  "Return from storage",
			CompanyName:           "Fingerprint",
			CompanyLocFromStorage: "Chicago, United States",
			ExpCompanyLocation:    "Chicago, United States",
		},
	}

	openaiKey := env.GetString(testutil.EnvKeyTestOpenAiKey)
	if !testing.Short() && openaiKey != "" {
		// test with using OpenAI API
		testCases = append(testCases, testCase{
			Name:               "Return from OpenAI",
			CompanyName:        "Haraj",
			ExpCompanyLocation: "Riyadh, Saudi Arabia",
		})
	}

	mockStorage := &mockStorage{}
	locator, err := hqloc.NewLocator(hqloc.LocatorConfig{
		Storage:       mockStorage,
		OpenaAiClient: openai.NewClient(option.WithAPIKey(openaiKey)),
	})
	require.NoError(t, err)

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			if testCase.CompanyLocFromStorage != "" {
				mockStorage.companyLocation = testCase.CompanyLocFromStorage
			} else {
				mockStorage.companyLocation = ""
			}

			loc, err := locator.Locate(context.Background(), testCase.CompanyName)
			require.NoError(t, err)
			require.Equal(t, testCase.ExpCompanyLocation, loc)
		})
	}
}

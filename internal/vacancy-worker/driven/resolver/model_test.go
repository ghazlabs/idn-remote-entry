package resolver_test

import (
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver"
	"github.com/stretchr/testify/assert"
)

func TestToVacancy(t *testing.T) {
	testCases := []struct {
		name     string
		input    resolver.VacancyInfo
		applyURL string
		want     *core.Vacancy
	}{
		{
			name: "basic vacancy info",
			input: resolver.VacancyInfo{
				JobTitle:         "Software Engineer",
				CompanyName:      "Example Corp",
				CompanyLocation:  "Global Remote",
				ShortDescription: "We are looking for a Software Engineer",
				RelevantTags:     []string{" remote", "software ", " engineer ", " ", ""},
			},
			applyURL: "https://example.com",
			want: &core.Vacancy{
				JobTitle:         "Software Engineer",
				CompanyName:      "Example Corp",
				CompanyLocation:  "Global Remote",
				ShortDescription: "We are looking for a Software Engineer",
				RelevantTags:     []string{"remote", "software", "engineer"},
				ApplyURL:         "https://example.com",
			},
		},
		{
			name: "empty tags",
			input: resolver.VacancyInfo{
				JobTitle:         "Product Manager",
				CompanyName:      "Tech Co",
				CompanyLocation:  "New York",
				ShortDescription: "Looking for PM",
				RelevantTags:     []string{},
			},
			applyURL: "https://techco.com",
			want: &core.Vacancy{
				JobTitle:         "Product Manager",
				CompanyName:      "Tech Co",
				CompanyLocation:  "New York",
				ShortDescription: "Looking for PM",
				RelevantTags:     []string{},
				ApplyURL:         "https://techco.com",
			},
		},
		{
			name: "nil tags",
			input: resolver.VacancyInfo{
				JobTitle:         "Product Manager",
				CompanyName:      "Tech Co",
				CompanyLocation:  "New York",
				ShortDescription: "Looking for PM",
				RelevantTags:     nil,
			},
			applyURL: "https://techco.com",
			want: &core.Vacancy{
				JobTitle:         "Product Manager",
				CompanyName:      "Tech Co",
				CompanyLocation:  "New York",
				ShortDescription: "Looking for PM",
				RelevantTags:     []string{},
				ApplyURL:         "https://techco.com",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.ToVacancy(tc.applyURL)
			assert.Equal(t, tc.want, got)
		})
	}
}

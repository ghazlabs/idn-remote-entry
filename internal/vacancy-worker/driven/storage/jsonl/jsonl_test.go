package jsonl

import (
	"context"
	"os"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/stretchr/testify/assert"
)

func TestJSONLStorageSave(t *testing.T) {
	filePath := "test_vacancies.jsonl"
	defer os.Remove(filePath)

	storage, err := NewJSONLStorage(JSONLStorageConfig{FilePath: filePath})
	assert.NoError(t, err)

	vacancy := core.Vacancy{
		JobTitle:         "Software Engineer",
		CompanyName:      "Tech Corp",
		CompanyLocation:  "Remote",
		ShortDescription: "Develop and maintain software applications.",
		RelevantTags:     []string{"Go", "Backend"},
		ApplyURL:         "https://example.com/apply",
	}

	record, err := storage.Save(context.Background(), vacancy)
	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, vacancy.JobTitle, record.Vacancy.JobTitle)
}

func TestJSONLStorageLookupCompanyLocation(t *testing.T) {
	filePath := "test_vacancies.jsonl"
	defer os.Remove(filePath)

	storage, err := NewJSONLStorage(JSONLStorageConfig{FilePath: filePath})
	assert.NoError(t, err)

	vacancy := core.Vacancy{
		JobTitle:         "Software Engineer",
		CompanyName:      "Tech Corp",
		CompanyLocation:  "New York",
		ShortDescription: "Develop and maintain software applications.",
		RelevantTags:     []string{"Go", "Backend"},
		ApplyURL:         "https://techcorp.com/apply",
	}

	_, err = storage.Save(context.Background(), vacancy)
	assert.NoError(t, err)

	location, err := storage.LookupCompanyLocation(context.Background(), "Tech Corp")
	assert.NoError(t, err)
	assert.Equal(t, "New York", location)
}

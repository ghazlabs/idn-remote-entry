package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv(testutil.EnvKeyMysqlDsn)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	// Clean up test data
	_, err = db.Exec("TRUNCATE TABLE vacancies")
	require.NoError(t, err)

	return db
}

func TestMySQLStorage_Save(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage, err := NewMySQLStorage(MySQLStorageConfig{DB: db})
	require.NoError(t, err)

	testCases := []struct {
		name     string
		vacancy  core.Vacancy
		validate func(*testing.T, *core.VacancyRecord)
	}{
		{
			name: "success",
			vacancy: core.Vacancy{
				JobTitle:         "Software Engineer",
				CompanyName:      "Test Company",
				CompanyLocation:  "Jakarta",
				ShortDescription: "Test Description",
				RelevantTags:     []string{"golang", "mysql"},
				ApplyURL:         "https://example.com/apply",
			},
			validate: func(t *testing.T, record *core.VacancyRecord) {
				assert.NotEmpty(t, record.ID)
				assert.Equal(t, "https://example.com/apply", record.PublicURL)
				assert.Equal(t, "Software Engineer", record.JobTitle)
				assert.Equal(t, "Test Company", record.CompanyName)
				assert.Equal(t, "Jakarta", record.CompanyLocation)
				assert.Equal(t, "Test Description", record.ShortDescription)
				assert.Equal(t, []string{"golang", "mysql"}, record.RelevantTags)
				assert.Equal(t, "https://example.com/apply", record.ApplyURL)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			record, err := storage.Save(context.Background(), tc.vacancy)
			assert.NoError(t, err)
			tc.validate(t, record)
		})
	}
}

func TestMySQLStorage_LookupCompanyLocation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	storage, err := NewMySQLStorage(MySQLStorageConfig{DB: db})
	require.NoError(t, err)

	// Insert test data
	vacancy := core.Vacancy{
		JobTitle:         "Software Engineer",
		CompanyName:      "Test Company",
		CompanyLocation:  "Jakarta",
		ShortDescription: "Test Description",
		RelevantTags:     []string{"golang", "mysql"},
		ApplyURL:         "https://example.com/apply",
	}
	_, err = storage.Save(context.Background(), vacancy)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		companyName string
		wantLoc     string
		wantErr     bool
	}{
		{
			name:        "existing company",
			companyName: "Test Company",
			wantLoc:     "Jakarta",
			wantErr:     false,
		},
		{
			name:        "non-existing company",
			companyName: "Unknown Company",
			wantLoc:     "",
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			location, err := storage.LookupCompanyLocation(context.Background(), tc.companyName)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantLoc, location)
		})
	}
}

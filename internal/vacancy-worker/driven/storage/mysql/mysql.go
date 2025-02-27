package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type MySQLStorage struct {
	MySQLStorageConfig
}

type MySQLStorageConfig struct {
	DB *sql.DB `validate:"nonnil"`
}

func NewMySQLStorage(cfg MySQLStorageConfig) (*MySQLStorage, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &MySQLStorage{
		MySQLStorageConfig: cfg,
	}, nil
}

func (s *MySQLStorage) Save(ctx context.Context, v core.Vacancy) (*core.VacancyRecord, error) {
	// Convert relevant_tags to JSON string
	relevantTagsJSON, err := json.Marshal(v.RelevantTags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal relevant tags: %w", err)
	}

	query := `
		INSERT INTO vacancies (
			job_title,
			company_name,
			company_location,
			short_description,
			relevant_tags,
			apply_url
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.DB.ExecContext(ctx, query,
		v.JobTitle,
		v.CompanyName,
		v.CompanyLocation,
		v.ShortDescription,
		relevantTagsJSON,
		v.ApplyURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save vacancy: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	rec := &core.VacancyRecord{
		ID:      fmt.Sprintf("%d", id),
		Vacancy: v,
		// Since in mysql there is no public URL, therefore return the apply url
		PublicURL: v.ApplyURL,
	}

	return rec, nil
}

func (s *MySQLStorage) LookupCompanyLocation(ctx context.Context, companyName string) (string, error) {
	query := `
		SELECT company_location FROM vacancies WHERE company_name = ? LIMIT 1
	`
	var location string
	err := s.DB.QueryRowContext(ctx, query, companyName).Scan(&location)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to lookup company location: %w", err)
	}

	lowerLocation := strings.ToLower(location)
	if strings.Contains(lowerLocation, "remote") || strings.Contains(lowerLocation, "global") {
		return "", nil
	}

	return location, nil
}

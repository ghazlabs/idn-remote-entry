package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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

package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"gopkg.in/validator.v2"
)

const (
	tableApproval = "approvals"
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

func (s *MySQLStorage) IsVacancyAlreadyRequested(ctx context.Context, applyURL string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM approvals
		WHERE JSON_EXTRACT(request_data, '$.apply_url') = ?
	`

	var count int
	err := s.DB.QueryRowContext(ctx, query, applyURL).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

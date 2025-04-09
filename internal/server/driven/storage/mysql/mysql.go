package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
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

func (s *MySQLStorage) GetApprovalState(ctx context.Context, messageID string) (core.ApprovalState, error) {
	var state string
	query := fmt.Sprintf("SELECT state FROM %s WHERE message_id = ?", tableApproval)
	err := s.DB.QueryRowContext(ctx, query, messageID).Scan(&state)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", shcore.NewBadRequestError("approval not found")
		}
		return "", fmt.Errorf("failed to get approval state: %w", err)
	}
	return core.ApprovalState(state), nil
}

func (s *MySQLStorage) UpdateApprovalState(ctx context.Context, messageID string, state core.ApprovalState) error {
	query := fmt.Sprintf("UPDATE %s SET state = ? WHERE message_id = ?", tableApproval)
	result, err := s.DB.ExecContext(ctx, query, state, messageID)
	if err != nil {
		return fmt.Errorf("failed to update approval state: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if affected == 0 {
		return shcore.NewBadRequestError(fmt.Sprintf("no approval request found with ID: %s", messageID))
	}

	return nil
}

func (s *MySQLStorage) SaveApprovalRequest(ctx context.Context, messageID string, req shcore.SubmitRequest) error {
	data := req.ToJSON()
	query := fmt.Sprintf("INSERT INTO %s (message_id, state, request_data) VALUES (?, ?, ?)", tableApproval)
	_, err := s.DB.ExecContext(ctx, query, messageID, core.ApprovalStatePending, data)
	if err != nil {
		return fmt.Errorf("failed to save approval request: %w", err)
	}
	return nil
}

func (s *MySQLStorage) SaveBulkApprovalRequest(ctx context.Context, req shcore.SubmitRequest, messageIDs []string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Prepare the query for bulk insertion
	query := fmt.Sprintf("INSERT INTO %s (message_id, state, request_data) VALUES (?, ?, ?)", tableApproval)
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Save each message ID with the corresponding request data
	for _, messageID := range messageIDs {
		data := req.ToJSON()
		_, err := stmt.ExecContext(ctx, messageID, core.ApprovalStatePending, data)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to save bulk approval request: %w", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

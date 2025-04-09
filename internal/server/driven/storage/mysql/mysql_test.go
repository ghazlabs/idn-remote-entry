package mysql

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	sharedcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/go-sql-driver/mysql"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := os.Getenv(testutil.EnvKeyMysqlDsn)
	db, err := sql.Open("mysql", dsn)
	require.NoError(t, err)

	// Clean up test data
	_, err = db.Exec("TRUNCATE TABLE approvals")
	require.NoError(t, err)

	return db
}

func TestGetApprovalState(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	storage, err := NewMySQLStorage(MySQLStorageConfig{DB: db})
	require.NoError(t, err)

	// Test getting non-existent approval
	state, err := storage.GetApprovalState(ctx, "non-existent")
	assert.Error(t, err)
	assert.Empty(t, state)

	// Insert test data
	messageID := "test-message"
	_, err = db.Exec("INSERT INTO approvals (message_id, state, request_data) VALUES (?, ?, '{}')",
		messageID, core.ApprovalStateApproved)
	require.NoError(t, err)

	// Test getting existing approval
	state, err = storage.GetApprovalState(ctx, messageID)
	assert.NoError(t, err)
	assert.Equal(t, core.ApprovalStateApproved, state)
}

func TestUpdateApprovalState(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	storage, err := NewMySQLStorage(MySQLStorageConfig{DB: db})
	require.NoError(t, err)

	// Test updating non-existent approval
	err = storage.UpdateApprovalState(ctx, "non-existent", core.ApprovalStateApproved)
	assert.Error(t, err)

	// Insert test data
	messageID := "test-message"
	_, err = db.Exec("INSERT INTO approvals (message_id, state, request_data) VALUES (?, ?, '{}')",
		messageID, core.ApprovalStatePending)
	require.NoError(t, err)

	// Test updating existing approval
	err = storage.UpdateApprovalState(ctx, messageID, core.ApprovalStateApproved)
	assert.NoError(t, err)

	// Verify state was updated
	var state string
	err = db.QueryRow("SELECT state FROM approvals WHERE message_id = ?", messageID).Scan(&state)
	require.NoError(t, err)
	assert.Equal(t, string(core.ApprovalStateApproved), state)
}

func TestSaveApprovalRequest(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	storage, err := NewMySQLStorage(MySQLStorageConfig{DB: db})
	require.NoError(t, err)

	messageID := "test-message"
	req := sharedcore.SubmitRequest{
		SubmissionType:  sharedcore.SubmitTypeManual,
		SubmissionEmail: "test@example.com",
		Vacancy: sharedcore.Vacancy{
			JobTitle:    "Software Engineer",
			CompanyName: "Test Company",
			ApplyURL:    "https://example.com/apply",
		},
	}

	// Test saving new approval request
	err = storage.SaveApprovalRequest(ctx, messageID, req)
	assert.NoError(t, err)

	// Verify request was saved
	var (
		state       string
		requestData []byte
	)
	err = db.QueryRow("SELECT state, request_data FROM approvals WHERE message_id = ?", messageID).
		Scan(&state, &requestData)
	require.NoError(t, err)
	assert.Equal(t, string(core.ApprovalStatePending), state)
	assert.JSONEq(t, string(req.ToJSON()), string(requestData))

	// Test saving duplicate request (should fail due to primary key)
	err = storage.SaveApprovalRequest(ctx, messageID, req)
	assert.Error(t, err)
}

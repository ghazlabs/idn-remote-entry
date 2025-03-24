package email

import (
	"context"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/assert"
)

func TestSendApprovalRequest(t *testing.T) {
	config := EmailConfig{
		Host:         env.GetString(testutil.EnvKeySMTPHost),
		Port:         env.GetInt(testutil.EnvKeySMTPPort),
		From:         "from@example.com",
		Password:     "password",
		ServerDomain: "http://example.com",
		AdminEmails:  "admin1@example.com,admin2@example.com",
	}

	client, err := NewEmail(config)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	req := core.SubmitRequest{
		SubmissionEmail: "user@example.com",
		SubmissionType:  core.SubmitTypeManual,
		Vacancy: core.Vacancy{
			JobTitle:         "Software Engineer",
			CompanyName:      "Example Inc.",
			CompanyLocation:  "Remote",
			ShortDescription: "We are looking for a Software Engineer.",
			RelevantTags:     []string{"Go", "Remote"},
			ApplyURL:         "http://example.com/apply",
		},
	}

	msgID, err := client.SendApprovalRequest(context.Background(), req, "token123")
	assert.NoError(t, err)
	assert.NotEmpty(t, msgID)
}

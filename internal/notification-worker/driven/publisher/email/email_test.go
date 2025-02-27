package email

import (
	"context"
	"testing"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/assert"
)

func TestEmailPublisher_Publish(t *testing.T) {
	publisher, err := NewEmailPublisher(EmailPublisherConfig{
		Host: env.GetString(testutil.EnvKeySMTPHost),
		Port: env.GetInt(testutil.EnvKeySMTPPort),
		From: "from@example.com",
		To:   "to@example.com",
	})
	assert.NoError(t, err)

	testCases := []struct {
		name         string
		notification shcore.Notification
		wantErr      bool
	}{
		{
			name: "valid notification",
			notification: shcore.Notification{
				VacancyRecord: shcore.VacancyRecord{
					Vacancy: shcore.Vacancy{
						JobTitle:         "Software Engineer",
						CompanyName:      "Test Company",
						CompanyLocation:  "Test Location",
						ShortDescription: "Test Description",
						RelevantTags:     []string{"golang", "testing"},
						ApplyURL:         "https://example.com/apply",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := publisher.Publish(context.Background(), tc.notification)
			assert.NoError(t, err)
		})
	}
}

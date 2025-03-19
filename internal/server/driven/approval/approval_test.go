package approval_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/approval"
)

func TestApproval_NeedsApproval(t *testing.T) {
	approvedEmails := "user1@example.com,user2@example.com,*@approved-domain.com"

	a, err := approval.NewApproval(approval.ApprovalConfig{
		ApprovedSubmitterEmails: approvedEmails,
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		submitterEmail string
		needsApproval  bool
	}{
		{
			name:           "Exact match email should not need approval",
			submitterEmail: "user1@example.com",
			needsApproval:  false,
		},
		{
			name:           "Domain match should not need approval",
			submitterEmail: "any.user@approved-domain.com",
			needsApproval:  false,
		},
		{
			name:           "Non-matching email should need approval",
			submitterEmail: "random@other-domain.com",
			needsApproval:  true,
		},
		{
			name:           "Partial domain match should need approval",
			submitterEmail: "user@sub.approved-domain.com",
			needsApproval:  true,
		},
		{
			name:           "Empty email should need approval",
			submitterEmail: "",
			needsApproval:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.NeedsApproval(tt.submitterEmail)
			assert.Equal(t, tt.needsApproval, got)
		})
	}
}

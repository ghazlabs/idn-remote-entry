package approval

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/validator.v2"
)

type ApprovalConfig struct {
	ApprovedSubmitterEmails string `validate:"nonzero"`
}

type Approval struct {
	ApprovalConfig
}

func NewApproval(cfg ApprovalConfig) (*Approval, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Approval{ApprovalConfig: cfg}, nil
}

func (a *Approval) NeedsApproval(submitterEmail string) bool {
	return !a.isSubmitterExempted(submitterEmail)
}

func (a *Approval) isSubmitterExempted(submitterEmail string) bool {
	if a.ApprovedSubmitterEmails == "" {
		return false
	}

	exemptedSubmitterEmailList := strings.Split(a.ApprovedSubmitterEmails, ",")

	return slices.ContainsFunc(exemptedSubmitterEmailList, func(emailPattern string) bool {
		emailPattern = strings.TrimSpace(emailPattern)
		if emailPattern == "" {
			return false
		}

		if strings.HasPrefix(emailPattern, "*@") {
			domain := strings.TrimPrefix(emailPattern, "*@")
			return strings.HasSuffix(submitterEmail, "@"+domain)
		}

		return emailPattern == submitterEmail
	})
}

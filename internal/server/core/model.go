package core

import shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"

type SubmitType string

const (
	SubmitTypeManual SubmitType = "manual"
	SubmitTypeURL    SubmitType = "url"
)

type SubmitRequest struct {
	SubmissionType SubmitType `json:"submission_type"`
	shcore.Vacancy
}

func (r SubmitRequest) Validate() error {
	// TODO
	return nil
}

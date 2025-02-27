package core

import "encoding/json"

type Vacancy struct {
	JobTitle         string   `json:"job_title,omitempty"`
	CompanyName      string   `json:"company_name,omitempty"`
	CompanyLocation  string   `json:"company_location,omitempty"`
	ShortDescription string   `json:"short_description,omitempty"`
	RelevantTags     []string `json:"relevant_tags,omitempty"`
	ApplyURL         string   `json:"apply_url,omitempty"`
}

type VacancyRecord struct {
	ID        string `json:"id"`
	PublicURL string `json:"public_url"`
	Vacancy
}

type Notification struct {
	RecipientID string `json:"recipient_id"`
	VacancyRecord
}

func (v Notification) ToJSON() []byte {
	data, _ := json.Marshal(v)
	return data
}

type SubmitType string

const (
	SubmitTypeManual SubmitType = "manual"
	SubmitTypeURL    SubmitType = "url"
)

type SubmitRequest struct {
	SubmissionType SubmitType `json:"submission_type"`
	Vacancy
}

func (r SubmitRequest) Validate() error {
	// TODO
	return nil
}

func (r SubmitRequest) ToJSON() []byte {
	data, _ := json.Marshal(r)
	return data
}

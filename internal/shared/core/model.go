package core

import "encoding/json"

type Vacancy struct {
	JobTitle         string   `json:"job_title"`
	CompanyName      string   `json:"company_name"`
	CompanyLocation  string   `json:"company_location"`
	ShortDescription string   `json:"short_description"`
	RelevantTags     []string `json:"relevant_tags"`
	ApplyURL         string   `json:"apply_url"`
}

type VacancyRecord struct {
	ID        string `json:"id"`
	PublicURL string `json:"public_url"`
	Vacancy
}

type WaNotification struct {
	RecipientID string `json:"recipient_id"`
	VacancyRecord
}

func (v WaNotification) ToJSON() []byte {
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

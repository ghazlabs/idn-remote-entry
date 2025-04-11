package core

import (
	"encoding/json"
	"fmt"

	"gopkg.in/validator.v2"
)

type Vacancy struct {
	JobTitle         string   `json:"job_title,omitempty" validate:"nonzero"`
	CompanyName      string   `json:"company_name,omitempty" validate:"nonzero"`
	CompanyLocation  string   `json:"company_location,omitempty"`
	ShortDescription string   `json:"short_description,omitempty"`
	RelevantTags     []string `json:"relevant_tags,omitempty"`
	ApplyURL         string   `json:"apply_url" validate:"nonzero"`
}

func (r Vacancy) Validate() error {
	err := validator.Validate(r)
	if err != nil {
		return fmt.Errorf("invalid vacancy: %w", err)
	}
	return nil
}

func (v Vacancy) ToJSON() []byte {
	data, _ := json.Marshal(v)
	return data
}

type VacancyRecord struct {
	ID        string `json:"id"`
	PublicURL string `json:"public_url"`
	Vacancy
}

type Notification struct {
	Retries int `json:"retries"`
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
	SubmitTypeBulk   SubmitType = "bulk"
)

type SubmitRequest struct {
	SubmissionType  SubmitType `json:"submission_type"`
	SubmissionEmail string     `json:"submission_email"`
	Retries         int        `json:"retries"`
	BulkVacancies   []Vacancy  `json:"bulk_vacancies,omitempty"`
	Vacancy
}

func (r SubmitRequest) Validate() error {
	if r.SubmissionType == SubmitTypeBulk {
		if len(r.BulkVacancies) == 0 {
			return fmt.Errorf("bulk_vacancies cannot be empty")
		}
		for _, v := range r.BulkVacancies {
			// Validate each vacancy in the bulk submission
			if err := v.Validate(); err != nil {
				return fmt.Errorf("invalid bulk vacancy: %w", err)
			}
		}
	}

	return nil
}

func (r SubmitRequest) ToJSON() []byte {
	data, _ := json.Marshal(r)
	return data
}

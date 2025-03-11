package core

import (
	"encoding/json"
	"fmt"

	"gopkg.in/validator.v2"
)

type Vacancy struct {
	JobTitle         string   `json:"job_title" validate:"nonzero"`
	CompanyName      string   `json:"company_name" validate:"nonzero"`
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

type VacancyRecord struct {
	ID        string `json:"id"`
	PublicURL string `json:"public_url"`
	Vacancy
}

type Notification struct {
	Retries int `json:"retries"`
	VacancyRecord
}

func (r Notification) GetRetries() int {
	return r.Retries
}

func (r *Notification) IncreaseRetries() {
	r.Retries++
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
	Retries        int        `json:"retries"`
	Vacancy
}

func (r SubmitRequest) GetRetries() int {
	return r.Retries
}

func (r *SubmitRequest) IncreaseRetries() {
	r.Retries++
}

func (r SubmitRequest) Validate() error {
	// TODO
	return nil
}

func (r SubmitRequest) ToJSON() []byte {
	data, _ := json.Marshal(r)
	return data
}

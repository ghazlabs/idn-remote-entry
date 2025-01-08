package core

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

type Vacancy struct {
	JobTitle         string   `json:"job_title"`
	CompanyName      string   `json:"company_name"`
	CompanyLocation  string   `json:"company_location"`
	ShortDescription string   `json:"short_description"`
	RelevantTags     []string `json:"relevant_tags"`
	ApplyURL         string   `json:"apply_url"`
}

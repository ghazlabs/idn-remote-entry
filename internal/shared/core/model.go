package core

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

package resolver

import (
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type VacancyInfo struct {
	JobTitle         string   `json:"job_title" jsonschema_description:"Title of the job taken from the vacancy description, if not found then it should be empty"`
	CompanyName      string   `json:"company_name" jsonschema_description:"Name of the company that posted the vacancy, if not found then it should be empty"`
	CompanyLocation  string   `json:"company_location" jsonschema_description:"Location of the company HQ based on the info found in the description, if not found then based on knowledge that you know, if still not found it should be filled with 'Global Remote'"`
	ShortDescription string   `json:"short_description" jsonschema_description:"Summary of the vacancy description, highlighting important points, if not found then it should be empty"`
	RelevantTags     []string `json:"relevant_tags" jsonschema_description:"List of tags that are relevant to the vacancy maximum 5 written in lowercase, if not found then it should be empty"`
}

func (i VacancyInfo) ToVacancy(url string) *core.Vacancy {
	tags := make([]string, 0)
	for _, tag := range i.RelevantTags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return &core.Vacancy{
		JobTitle:         i.JobTitle,
		CompanyName:      i.CompanyName,
		CompanyLocation:  i.CompanyLocation,
		ShortDescription: i.ShortDescription,
		RelevantTags:     tags,
		ApplyURL:         url,
	}
}

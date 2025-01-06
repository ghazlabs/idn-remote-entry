package resolver

import "github.com/invopop/jsonschema"

type vacancyInfo struct {
	JobTitle         string   `json:"job_title" jsonschema_description:"Title of the job taken from the vacancy description"`
	CompanyName      string   `json:"company_name" jsonschema_description:"Name of the company that posted the vacancy"`
	ShortDescription string   `json:"short_description" jsonschema_description:"Summary of the vacancy description, highlighting important points"`
	RelevantTags     []string `json:"relevant_tags" jsonschema_description:"List of tags that are relevant to the vacancy maximum 5 written in lowercase"`
}

type locationInfo struct {
	Location string `json:"country" jsonschema_description:"Location where the company HQ is located, it should be in the format of City, Country"`
}

func generateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

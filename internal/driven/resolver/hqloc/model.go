package hqloc

import "strings"

type NotionResponse struct {
	Results []Page `json:"results"`
}

func (r NotionResponse) GetCompanyLocation() string {
	if len(r.Results) == 0 {
		return ""
	}
	compLoc := r.Results[0].Properties.CompanyLocation.RichText[0].Text.Content
	lowerCompLoc := strings.ToLower(compLoc)
	if strings.Contains(lowerCompLoc, "remote") || strings.Contains(lowerCompLoc, "global") {
		return ""
	}

	return compLoc
}

type Page struct {
	Properties Properties `json:"properties"`
}

type Properties struct {
	CompanyLocation CompanyLocation `json:"Company Location"`
}

type CompanyLocation struct {
	RichText []RichText `json:"rich_text"`
}

type RichText struct {
	Text TextContent `json:"text"`
}

type TextContent struct {
	Content string `json:"content"`
}

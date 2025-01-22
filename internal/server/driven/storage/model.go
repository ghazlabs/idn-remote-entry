package storage

import (
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
)

type BlockTitle struct {
	Title []BlockTextContent `json:"title"`
}

type BlockRichText struct {
	RichText []BlockTextContent `json:"rich_text"`
}

type BlockTextContent struct {
	Text struct {
		Content string `json:"content"`
	} `json:"text"`
}

type BlockURL struct {
	URL string `json:"url"`
}

type BlockDate struct {
	Date struct {
		Start string `json:"start"`
	} `json:"date"`
}

type BlockMultiSelect struct {
	MultiSelect []BlockSelect `json:"multi_select"`
}

type BlockSelect struct {
	Name string `json:"name"`
}

type InsertRecordPayload struct {
	Parent struct {
		DatabaseID string `json:"database_id"`
	} `json:"parent"`
	Properties struct {
		Title            BlockTitle       `json:"Title"`
		CompanyName      BlockRichText    `json:"Company Name"`
		CompanyLocation  BlockRichText    `json:"Company Location"`
		ShortDescription BlockRichText    `json:"Short Description"`
		RelevantTags     BlockMultiSelect `json:"Relevant Tags"`
		ApplyURL         BlockURL         `json:"Apply URL"`
		DateAdded        BlockDate        `json:"Date Added"`
	} `json:"properties"`
}

func NewInsertRecordPaylod(databaseID string, now time.Time, v core.Vacancy) InsertRecordPayload {
	var p InsertRecordPayload

	p.Parent.DatabaseID = databaseID

	var title BlockTextContent
	title.Text.Content = v.JobTitle
	p.Properties.Title.Title = []BlockTextContent{title}

	var companyName BlockTextContent
	companyName.Text.Content = v.CompanyName
	p.Properties.CompanyName.RichText = []BlockTextContent{companyName}

	var companyLocation BlockTextContent
	companyLocation.Text.Content = v.CompanyLocation
	p.Properties.CompanyLocation.RichText = []BlockTextContent{companyLocation}

	var shortDescription BlockTextContent
	shortDescription.Text.Content = v.ShortDescription
	p.Properties.ShortDescription.RichText = []BlockTextContent{shortDescription}

	var selects []BlockSelect
	for _, tag := range v.RelevantTags {
		var selectItem BlockSelect
		selectItem.Name = tag
		selects = append(selects, selectItem)
	}
	p.Properties.RelevantTags.MultiSelect = selects

	var applyURL BlockURL
	applyURL.URL = v.ApplyURL
	p.Properties.ApplyURL = applyURL

	var dateAdded BlockDate
	dateAdded.Date.Start = now.Format("2006-01-02")
	p.Properties.DateAdded = dateAdded

	return p
}

type insertRecordResponse struct {
	ID        string `json:"id"`
	PublicURL string `json:"public_url"`
}

package notion

import (
	"context"
	"fmt"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

const (
	// TODO: move this to config
	URIIDNRemote = "https://idnremote.com/#/jobs/"
)

type NotionStorage struct {
	NotionStorageConfig
}

type NotionStorageConfig struct {
	DatabaseID  string        `validate:"nonzero"`
	NotionToken string        `validate:"nonzero"`
	HttpClient  *resty.Client `validate:"nonnil"`
}

func NewNotionStorage(cfg NotionStorageConfig) (*NotionStorage, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &NotionStorage{
		NotionStorageConfig: cfg,
	}, nil
}

func (s *NotionStorage) Save(ctx context.Context, v core.Vacancy) (*core.VacancyRecord, error) {
	var respBody insertRecordResponse
	payload := NewInsertRecordPaylod(s.DatabaseID, time.Now(), v)
	resp, err := s.HttpClient.R().
		SetHeader("Authorization", "Bearer "+s.NotionToken).
		SetHeader("Content-Type", "application/json").
		SetHeader("Notion-Version", "2022-06-28").
		SetBody(payload).
		SetResult(&respBody).
		Post("https://api.notion.com/v1/pages")
	if err != nil {
		return nil, fmt.Errorf("failed to call api to save the vacancy: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("failed to save the vacancy: %s", resp.String())
	}

	rec := &core.VacancyRecord{
		ID:        respBody.ID,
		Vacancy:   v,
		PublicURL: fmt.Sprintf("%s%s", URIIDNRemote, respBody.ID),
	}
	return rec, nil
}

func (s *NotionStorage) LookupCompanyLocation(ctx context.Context, companyName string) (string, error) {
	var respBody lookupRecordResponse
	resp, err := s.HttpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+s.NotionToken).
		SetHeader("Content-Type", "application/json").
		SetHeader("Notion-Version", "2022-06-28").
		SetBody(map[string]interface{}{
			"filter": map[string]interface{}{
				"property": "Company Name",
				"rich_text": map[string]string{
					"equals": companyName,
				},
			},
			"sorts": []map[string]string{
				{
					"property":  "Last edited time",
					"direction": "descending",
				},
			},
			"page_size": 1,
		}).
		SetResult(&respBody).
		Post(fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", s.DatabaseID))
	if err != nil {
		return "", fmt.Errorf("failed to call notion api to lookup company location: %w", err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("failed to lookup company location from Notion: %s", resp.String())
	}

	return respBody.GetCompanyLocation(), nil
}

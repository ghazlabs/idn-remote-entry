package notion

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
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

func (s *NotionStorage) GetAllURLVacancies(ctx context.Context) (map[string]bool, error) {
	var allPages []Page
	var respBody lookupRecordResponse
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", s.DatabaseID)
	nextCursor := ""

	for {
		body := map[string]interface{}{}
		if nextCursor != "" {
			body["start_cursor"] = nextCursor
		}

		resp, err := s.HttpClient.R().
			SetContext(ctx).
			SetHeader("Authorization", "Bearer "+s.NotionToken).
			SetHeader("Content-Type", "application/json").
			SetHeader("Notion-Version", "2022-06-28").
			SetResult(&respBody).
			SetBody(body).
			Post(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch pages from Notion: %w", err)
		}
		if resp.IsError() {
			return nil, fmt.Errorf("failed to fetch pages from Notion: %s", resp.String())
		}

		allPages = append(allPages, respBody.Results...)

		if respBody.HasMore == false {
			break
		}
		nextCursor = respBody.NextCursor
	}

	vacancies := make(map[string]bool)
	for _, page := range allPages {
		vacancies[page.Properties.ApplyURL.URL] = true
	}

	return vacancies, nil
}

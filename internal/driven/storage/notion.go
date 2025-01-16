package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/core"
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
		PublicURL: respBody.PublicURL,
	}
	return rec, nil
}

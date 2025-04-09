package server

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

type Server struct {
	ServerConfig
}

type ServerConfig struct {
	HttpClient *resty.Client `validate:"nonnil"`
	ApiKey     string        `validate:"nonzero"`
}

func NewClientServer(cfg ServerConfig) (*Server, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &Server{
		ServerConfig: cfg,
	}, nil
}

func (s *Server) SubmitBulkVacancies(ctx context.Context, vacancies []core.Vacancy) error {
	payload := map[string]interface{}{
		"submission_type":  "bulk",
		"submission_email": "crawler",
		"bulk_vacancies":   vacancies,
	}
	resp, err := s.HttpClient.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Api-Key", s.ApiKey).
		SetBody(payload).
		Post("/vacancies")
	if err != nil {
		return fmt.Errorf("failed to call api to submit bulk vacancies: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to submit bulk vacancies: %s", resp.String())
	}

	return nil
}

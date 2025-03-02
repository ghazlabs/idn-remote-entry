package jsonl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type JSONLStorage struct {
	filePath string
}

type JSONLStorageConfig struct {
	FilePath string `validate:"nonzero"`
}

func NewJSONLStorage(cfg JSONLStorageConfig) (*JSONLStorage, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// check if the file exists
	if _, err := os.Stat(cfg.FilePath); os.IsNotExist(err) {
		// create the file
		file, err := os.Create(cfg.FilePath)
		defer file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to create JSONL file: %w", err)
		}
	}

	return &JSONLStorage{
		filePath: cfg.FilePath,
	}, nil
}

func (s *JSONLStorage) Save(ctx context.Context, v core.Vacancy) (*core.VacancyRecord, error) {
	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	record := core.VacancyRecord{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Vacancy:   v,
		PublicURL: v.ApplyURL,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record: %w", err)
	}

	if _, err := file.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to file: %w", err)
	}

	if _, err := file.WriteString("\n"); err != nil {
		return nil, fmt.Errorf("failed to write newline to file: %w", err)
	}

	return &record, nil
}

func (s *JSONLStorage) LookupCompanyLocation(ctx context.Context, companyName string) (string, error) {
	file, err := os.Open(s.filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var record core.VacancyRecord
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to decode record: %w", err)
		}

		if record.Vacancy.CompanyName == companyName {
			lowerLocation := strings.ToLower(record.Vacancy.CompanyLocation)
			if strings.Contains(lowerLocation, "remote") || strings.Contains(lowerLocation, "global") {
				return "", nil
			}
			return record.Vacancy.CompanyLocation, nil
		}
	}

	return "", nil
}

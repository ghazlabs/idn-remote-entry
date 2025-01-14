package parser

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

type TextParser struct {
	TextParserConfig
}

type TextParserConfig struct {
	HttpClient *resty.Client `validate:"nonnil"`
}

func NewTextParser(cfg TextParserConfig) (*TextParser, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &TextParser{
		TextParserConfig: cfg,
	}, nil
}

func (p *TextParser) GetText(ctx context.Context, url string) (string, error) {
	// get the html content of the URL
	resp, err := p.HttpClient.R().SetContext(ctx).Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to open the URL: %w", err)
	}

	// parse the HTML content
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return "", fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	// Remove <script> and <style> tags from the document
	doc.Find("script, style").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	return strings.TrimSpace(doc.Find("body").Text()), nil
}

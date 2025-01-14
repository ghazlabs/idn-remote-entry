package parser

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type TextParser struct {
	TextParserConfig
}

type TextParserConfig struct {
	HttpClient    *resty.Client  `validate:"nonnil"`
	OpenaAiClient *openai.Client `validate:"nonnil"`
}

func NewTextParser(cfg TextParserConfig) (*TextParser, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &TextParser{
		TextParserConfig: cfg,
	}, nil
}

func (p *TextParser) Parse(ctx context.Context, url string) (*core.Vacancy, error) {
	text, err := p.GetText(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get text content: %w", err)
	}

	// call the OpenAI API to parse the vacancy information
	vacInfo, err := resolver.CallOpenAI[resolver.VacancyInfo](ctx, p.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("I will give you unstructured text content of a remote vacancy, and you need to parse information from this text."),
		openai.UserMessage(text),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	return vacInfo.ToVacancy(), nil
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

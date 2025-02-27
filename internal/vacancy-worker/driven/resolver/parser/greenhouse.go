package parser

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/util"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type GreenhouseParser struct {
	GreenhouseParserConfig
}

type GreenhouseParserConfig struct {
	HttpClient    *resty.Client  `validate:"nonnil"`
	OpenaAiClient *openai.Client `validate:"nonnil"`
	ModelLLM      string         `validate:"nonzero"`
}

func NewGreenhouseParser(cfg GreenhouseParserConfig) (*GreenhouseParser, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &GreenhouseParser{
		GreenhouseParserConfig: cfg,
	}, nil
}

func (p *GreenhouseParser) Parse(ctx context.Context, url string) (*core.Vacancy, error) {
	text, jobTitle, err := p.getInfo(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get text content: %w", err)
	}

	// call the OpenAI API to parse the vacancy information
	vacInfo, err := util.CallOpenAI[resolver.VacancyInfo](ctx, p.OpenaAiClient, p.ModelLLM, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("I will give you unstructured text content of a remote vacancy, and you need to parse information from this text. Also please provide the company HQ location based on what you know."),
		openai.UserMessage(text),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	vac := vacInfo.ToVacancy(url)
	vac.JobTitle = jobTitle

	return vac, nil
}

func (p *GreenhouseParser) getInfo(ctx context.Context, url string) (string, string, error) {
	// get the html content of the URL
	resp, err := p.HttpClient.R().SetContext(ctx).Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to open the URL: %w", err)
	}

	// parse the HTML content
	respBody := bytes.NewReader(resp.Body())
	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	// Remove <script> and <style> tags from the document
	doc.Find("script, style").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	textBody := strings.TrimSpace(doc.Find("body").Text())
	metaTag := doc.Find(`meta[property="og:title"]`)
	jobTitle, _ := metaTag.Attr("content")

	return textBody, jobTitle, nil
}

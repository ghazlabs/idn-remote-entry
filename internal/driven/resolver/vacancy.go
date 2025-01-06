package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type VacancyResolver struct {
	VacancyResolverConfig
}

type VacancyResolverConfig struct {
	HTTPClient    *resty.Client  `validate:"nonnil"`
	OpenaAiClient *openai.Client `validate:"nonnil"`
}

func NewVacancyResolver(cfg VacancyResolverConfig) (*VacancyResolver, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &VacancyResolver{
		VacancyResolverConfig: cfg,
	}, nil
}

func (r *VacancyResolver) Resolve(ctx context.Context, url string) (*core.Vacancy, error) {
	// get the text content of the URL
	urlContent, err := r.getTextContent(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get the text content of the URL: %w", err)
	}
	log.Printf("URL Content: %s", urlContent)

	// call the OpenAI API to parse the vacancy information
	vacInfo, err := callOpenAI[vacancyInfo](ctx, r.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You will be given vacancy description and you need to parse the information from it."),
		openai.UserMessage(urlContent),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	// lookup to google to get the company HQ location
	textLocation, err := r.getTextContent(ctx, fmt.Sprintf("https://duckduckgo.com/html/?q=%s HQ Location", vacInfo.CompanyName))
	if err != nil {
		return nil, fmt.Errorf("failed to get reference of the company HQ location: %w", err)
	}

	// call the OpenAI API to get the company HQ location
	locInfo, err := callOpenAI[locationInfo](ctx, r.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You will be given a text content for a company location and you need to parse the location information from it."),
		openai.UserMessage(textLocation),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the location information: %w", err)
	}

	vac := &core.Vacancy{
		JobTitle:         vacInfo.JobTitle,
		CompanyName:      vacInfo.CompanyName,
		CompanyLocation:  locInfo.Location,
		ShortDescription: vacInfo.ShortDescription,
		RelevantTags:     vacInfo.RelevantTags,
		ApplyURL:         url,
	}

	return vac, nil
}

func (r *VacancyResolver) getTextContent(ctx context.Context, url string) (string, error) {
	// get the html content of the URL
	resp, err := r.HTTPClient.R().SetContext(ctx).Get(url)
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

func callOpenAI[T any](
	ctx context.Context,
	client *openai.Client,
	msgs []openai.ChatCompletionMessageParamUnion,
) (*T, error) {
	// generate schema for the structured output
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.String("info"),
		Description: openai.String("info"),
		Schema:      openai.F(generateSchema[T]()),
		Strict:      openai.Bool(true),
	}

	// call the OpenAI API
	chat, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(msgs),
		ResponseFormat: openai.F[openai.ChatCompletionNewParamsResponseFormatUnion](
			openai.ResponseFormatJSONSchemaParam{
				Type:       openai.F(openai.ResponseFormatJSONSchemaTypeJSONSchema),
				JSONSchema: openai.F(schemaParam),
			},
		),
		Model: openai.String(openai.ChatModelGPT4o2024_08_06),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call the OpenAI API: %w", err)
	}

	var info T
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &info)
	if err != nil {
		log.Fatalf("failed to unmarshal vacancy info: %v", err)
	}

	return &info, nil
}
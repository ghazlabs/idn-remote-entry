package resolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type VacancyResolver struct {
	VacancyResolverConfig
}

type VacancyResolverConfig struct {
	HttpClient    *resty.Client  `validate:"nonnil"`
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
	screenshot, err := r.takeScreenshot(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to take the screenshot of the URL: %w", err)
	}

	textContent, err := r.doOCR(ctx, screenshot)
	if err != nil {
		return nil, fmt.Errorf("failed to do OCR: %w", err)
	}

	// call the OpenAI API to parse the vacancy information
	vacInfo, err := callOpenAI[vacancyInfo](ctx, r.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("I will give you unstructured text content of a remote vacancy, and you need to parse information from this text."),
		openai.UserMessage(textContent),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	vac := &core.Vacancy{
		JobTitle:         vacInfo.JobTitle,
		CompanyName:      vacInfo.CompanyName,
		CompanyLocation:  vacInfo.CompanyLocation,
		ShortDescription: vacInfo.ShortDescription,
		RelevantTags:     vacInfo.RelevantTags,
		ApplyURL:         url,
	}

	return vac, nil
}

func (r *VacancyResolver) takeScreenshot(ctx context.Context, url string) ([]byte, error) {
	// create context for chrome
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// allocate a buffer to store the screenshot
	var buf []byte

	// capture the screenshot
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to take the screenshot: %w", err)
	}

	return buf, nil
}

type tesseractServerResponse struct {
	Data struct {
		Stdout string `json:"stdout"`
	} `json:"data"`
}

func (r *VacancyResolver) doOCR(ctx context.Context, buf []byte) (string, error) {
	// call tesseract server to do OCR
	var serverResp tesseractServerResponse
	_, err := r.HttpClient.R().
		SetContext(ctx).
		SetFileReader("file", "file", bytes.NewReader(buf)).
		SetFormData(map[string]string{
			"options": `{"languages": ["eng"]}`, // Add the "options" field
		}).
		SetResult(&serverResp).
		Post("http://127.0.0.1:8884/tesseract")
	if err != nil {
		return "", fmt.Errorf("failed to call the OCR server: %w", err)
	}
	return serverResp.Data.Stdout, nil
}

func (r *VacancyResolver) getTextContent(ctx context.Context, url string) (string, error) {
	// get the html content of the URL
	resp, err := r.HttpClient.R().SetContext(ctx).Get(url)
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
		Temperature: openai.Float(0.0),
		Model:       openai.String(openai.ChatModelGPT4o2024_08_06),
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

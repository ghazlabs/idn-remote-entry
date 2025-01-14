package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type VacancyResolver struct {
	VacancyResolverConfig
}

type Parser interface {
	GetText(ctx context.Context, url string) (string, error)
}

type ParserRegistry struct {
	ApexDomains []string
	Parser      Parser
}

type VacancyResolverConfig struct {
	OpenaAiClient    *openai.Client `validate:"nonnil"`
	DefaultParser    Parser         `validate:"nonnil"`
	ParserRegistries []ParserRegistry
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
	var textContent string
	var err error
	for _, reg := range r.ParserRegistries {
		for _, apex := range reg.ApexDomains {
			if strings.Contains(url, apex) {
				textContent, err = reg.Parser.GetText(ctx, url)
				if err != nil {
					return nil, fmt.Errorf("failed to get text content: %w", err)
				}
				goto parserFound
			}
		}
	}
	textContent, err = r.DefaultParser.GetText(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to get text content: %w", err)
	}

parserFound:
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

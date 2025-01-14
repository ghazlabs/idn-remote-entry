package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
)

func GenerateSchema[T any]() interface{} {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

func CallOpenAI[T any](
	ctx context.Context,
	client *openai.Client,
	msgs []openai.ChatCompletionMessageParamUnion,
) (*T, error) {
	// generate schema for the structured output
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        openai.String("info"),
		Description: openai.String("info"),
		Schema:      openai.F(GenerateSchema[T]()),
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

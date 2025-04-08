package contentchecker

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/util"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type applicable struct {
	IsApplicable bool `json:"is_applicable" jsonschema_description:"Is the vacancy applicable to apply."`
}

type Checker struct {
	CheckerConfig
}

type CheckerConfig struct {
	OpenaAiClient *openai.Client `validate:"nonnil"`
}

func NewContentChecker(cfg CheckerConfig) (*Checker, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &Checker{
		CheckerConfig: cfg,
	}, nil
}

func (c *Checker) IsApplicable(ctx context.Context, v core.Vacancy) (bool, error) {
	applicable, err := util.CallOpenAI[applicable](ctx, c.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a job vacancy checker that check its description if the vacancy is applicable to apply or not. The vacancy should be in engineering like software engineer or creative job such as designer. The job needs to be full remote or applicable for Indonesian talent that live in timezone GMT+7 and GMT+8. If you are unsure about the vacancy, consider it as not applicable."),
		openai.UserMessage(v.ShortDescription),
	})
	if err != nil {
		return false, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	return applicable.IsApplicable, nil
}

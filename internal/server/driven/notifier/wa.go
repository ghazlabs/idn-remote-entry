package notifier

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

type WhatsappNotifier struct {
	WhatsappNotifierConfig
}

func NewWhatsappNotifier(cfg WhatsappNotifierConfig) (*WhatsappNotifier, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &WhatsappNotifier{
		WhatsappNotifierConfig: cfg,
	}, nil
}

type WhatsappNotifierConfig struct {
	HttpClient           *resty.Client `validate:"nonnil"`
	Username             string        `validate:"nonzero"`
	Password             string        `validate:"nonzero"`
	WhatsappRecipientIDs []string      `validate:"nonzero"`
}

func (n *WhatsappNotifier) Notify(ctx context.Context, v core.VacancyRecord) error {
	for _, waID := range n.WhatsappRecipientIDs {
		err := n.notifyRecipient(ctx, v, waID)
		if err != nil {
			return fmt.Errorf("failed to notify recipient: %w", err)
		}
	}

	return nil
}

func (n *WhatsappNotifier) notifyRecipient(ctx context.Context, v core.VacancyRecord, waID string) error {
	// send notification to whatsapp using Ghazlabs Whatsapp API
	resp, err := n.HttpClient.R().
		SetContext(ctx).
		SetBasicAuth(n.Username, n.Password).
		SetBody(map[string]interface{}{
			"phone":   waID,
			"message": convertVacancyToMessage(v),
		}).
		Post("https://wa.ghazlabs.com/send/message")
	if err != nil {
		return fmt.Errorf("unable to make http request: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to send notification: %s", resp.String())
	}

	return nil
}

func convertVacancyToMessage(v core.VacancyRecord) string {
	tags := []string{}
	for _, tag := range v.RelevantTags {
		tags = append(tags, fmt.Sprintf("#%v", strings.ReplaceAll(tag, " ", "-")))
	}
	content := []string{
		fmt.Sprintf("✨ *%v*", strings.ToUpper(v.JobTitle)),
		"",
		fmt.Sprintf("🏢 %v", v.CompanyName),
		fmt.Sprintf("📍 %v", v.CompanyLocation),
		// "",
		// v.ShortDescription,
		"",
		fmt.Sprintf("%v", v.PublicURL),
		"",
		strings.Join(tags, " "),
	}
	return strings.Join(content, "\n")
}

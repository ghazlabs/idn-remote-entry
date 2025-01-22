package notifier

import (
	"context"
	"fmt"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

type WaPublisher struct {
	WaPublisherConfig
}

func NewWaPublisher(cfg WaPublisherConfig) (*WaPublisher, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &WaPublisher{
		WaPublisherConfig: cfg,
	}, nil
}

type WaPublisherConfig struct {
	HttpClient           *resty.Client `validate:"nonnil"`
	Username             string        `validate:"nonzero"`
	Password             string        `validate:"nonzero"`
	WhatsappRecipientIDs []string      `validate:"nonzero"`
	WaServerURL          string        `validate:"nonzero"`
}

func (n *WaPublisher) Publish(ctx context.Context, ntf core.WhatsappNotification) error {
	// send notification to whatsapp using Ghazlabs Whatsapp API
	resp, err := n.HttpClient.R().
		SetContext(ctx).
		SetBasicAuth(n.Username, n.Password).
		SetBody(map[string]interface{}{
			"phone":   ntf.RecipientID,
			"message": convertVacancyToMessage(ntf.VacancyRecord),
		}).
		Post(fmt.Sprintf("%v/send/message", n.WaServerURL))
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

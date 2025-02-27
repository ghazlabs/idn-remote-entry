package email

import (
	"context"
	"fmt"
	"net/smtp"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type EmailPublisherConfig struct {
	Host string `validate:"nonzero"`
	Port int    `validate:"nonzero"`
	From string `validate:"nonzero"`
	To   string `validate:"nonzero"`
}

type EmailPublisher struct {
	EmailPublisherConfig
}

func NewEmailPublisher(config EmailPublisherConfig) (*EmailPublisher, error) {
	err := validator.Validate(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &EmailPublisher{
		EmailPublisherConfig: config,
	}, nil
}

func (p *EmailPublisher) Publish(ctx context.Context, notification shcore.Notification) error {
	addr := fmt.Sprintf("%s:%d", p.Host, p.Port)

	subject := "New Job Vacancy Notification"
	body := fmt.Sprintf(`
Job Title: %s
Company: %s
Location: %s
Description: %s
URL: %s
`,
		notification.VacancyRecord.Vacancy.JobTitle,
		notification.VacancyRecord.Vacancy.CompanyName,
		notification.VacancyRecord.Vacancy.CompanyLocation,
		notification.VacancyRecord.Vacancy.ShortDescription,
		notification.VacancyRecord.Vacancy.ApplyURL,
	)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", p.From, p.To, subject, body)

	return smtp.SendMail(
		addr,
		nil, // no auth for local
		p.From,
		[]string{p.To},
		[]byte(msg),
	)
}

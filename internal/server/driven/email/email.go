package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type EmailConfig struct {
	Host         string `validate:"nonzero"`
	Port         int    `validate:"nonzero"`
	From         string `validate:"nonzero"`
	Password     string `validate:"nonzero"`
	ServerDomain string `validate:"nonzero"`
	AdminEmails  string `validate:"nonzero"`
}

type EmailClient struct {
	EmailConfig
}

func NewEmail(config EmailConfig) (*EmailClient, error) {
	if err := validator.Validate(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &EmailClient{EmailConfig: config}, nil
}

func (e *EmailClient) SendApprovalRequest(ctx context.Context, req core.SubmitRequest, tokenReq string) (string, error) {
	messageID := generateMessageID("idnremote.com")
	codeID := getCodeMessageID(messageID)
	headers := e.buildHeaders(messageID, "", fmt.Sprintf("IDNRemote.com - New Job Vacancy Approval - ID: %s", codeID))

	body := templateEmail{
		req:          req,
		token:        tokenReq,
		messageID:    messageID,
		serverDomain: e.ServerDomain,
	}

	if err := e.sendEmail(headers, messageID, body.getContentBodyHTML(), body.getContentBodyPlain()); err != nil {
		return "", err
	}

	return messageID, nil
}

func (e *EmailClient) SendBulkApprovalRequest(ctx context.Context, req core.SubmitRequest, tokenReqVacancies []string) ([]string, error) {
	messageID := generateMessageID("bulk")
	codeID := getCodeMessageID(messageID)
	headers := e.buildHeaders(messageID, "", fmt.Sprintf("IDNRemote.com - New Crawled Job Vacancy Approval - ID: %s", codeID))

	messagesIDs := make([]string, len(req.BulkVacancies))
	for i, _ := range req.BulkVacancies {
		messagesIDs[i] = generateMessageID("bulk")
	}

	body := templateBulkEmail{
		req:            req,
		tokenVacancies: tokenReqVacancies,
		messagesIDs:    messagesIDs,
		serverDomain:   e.ServerDomain,
	}

	if err := e.sendEmail(headers, messageID, body.getContentBodyHTML(), body.getContentBodyPlain()); err != nil {
		return nil, err
	}

	return messagesIDs, nil
}

func (e *EmailClient) ApproveRequest(ctx context.Context, messageID string) error {
	message := "Approved by admin"
	codeID := getCodeMessageID(messageID)
	headers := e.buildHeaders(messageID, messageID, fmt.Sprintf("Re: IDNRemote.com - New Job Vacancy Approval - ID: %s", codeID))
	return e.sendEmail(headers, messageID, message, message)
}

func (e *EmailClient) RejectRequest(ctx context.Context, messageID string) error {
	message := "Rejected by admin"
	codeID := getCodeMessageID(messageID)
	headers := e.buildHeaders(messageID, messageID, fmt.Sprintf("Re: IDNRemote.com - New Job Vacancy Approval - ID: %s", codeID))
	return e.sendEmail(headers, messageID, message, message)
}

func (e *EmailClient) buildHeaders(messageID, inReplyTo, subject string) map[string]string {
	headers := map[string]string{
		"From":    fmt.Sprintf("IDN Remote Entry <%s>", e.From),
		"To":      strings.Join(strings.Split(e.AdminEmails, ","), ", "),
		"Subject": subject,
	}

	if messageID != "" {
		headers["Message-ID"] = messageID
	}
	if inReplyTo != "" {
		headers["In-Reply-To"] = inReplyTo
		headers["References"] = inReplyTo
	}

	return headers
}

func (e *EmailClient) sendEmail(headers map[string]string, messageID, bodyHTML, bodyPlain string) error {
	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)
	emailReceivers := strings.Split(e.AdminEmails, ",")
	boundary := getCodeMessageID(messageID)

	message := e.buildEmailMessage(headers, boundary, bodyHTML, bodyPlain)

	var auth smtp.Auth
	if !e.isLocal() {
		auth = smtp.PlainAuth("IDN Remote Entry", e.From, e.Password, e.Host)
	}

	return smtp.SendMail(addr, auth, e.From, emailReceivers, []byte(message))
}

func (e *EmailClient) buildEmailMessage(headers map[string]string, boundary, bodyHTML, bodyPlain string) string {
	var message strings.Builder

	// Add headers
	for k, v := range headers {
		fmt.Fprintf(&message, "%s: %s\r\n", k, v)
	}

	// Add MIME headers
	fmt.Fprintf(&message, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&message, "Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary)

	// Add plain text version
	fmt.Fprintf(&message, "--%s\r\n", boundary)
	fmt.Fprintf(&message, "Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(&message, "%s\r\n\r\n", strings.ReplaceAll(bodyPlain, "<br>", "\n"))

	// Add HTML version
	fmt.Fprintf(&message, "--%s\r\n", boundary)
	fmt.Fprintf(&message, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	fmt.Fprintf(&message, `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body>
    %s
</body>
</html>`, bodyHTML)

	fmt.Fprintf(&message, "\r\n\r\n--%s--", boundary)
	return message.String()
}

func (e EmailClient) isLocal() bool {
	return e.Host == "mailpit" || strings.Contains(e.Host, "localhost")
}

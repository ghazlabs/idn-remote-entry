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
	err := validator.Validate(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &EmailClient{
		EmailConfig: config,
	}, nil
}

func (e *EmailClient) SendApprovalRequest(
	ctx context.Context,
	req core.SubmitRequest,
	tokenReq string,
) error {
	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)
	emailReceiverList := strings.Split(e.AdminEmails, ",")
	subject := "Action Required - New Job Vacancy Approval"
	body := e.setEmailApprovalBody(req, tokenReq)
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", e.From, e.AdminEmails, subject, body)

	var auth smtp.Auth
	if e.Host == "mailpit" {
		auth = nil
	} else {
		auth = smtp.PlainAuth("IDN Remote Entry", e.From, e.Password, e.Host)
	}

	return smtp.SendMail(
		addr,
		auth,
		e.From,
		emailReceiverList,
		[]byte(msg),
	)
}

func (e *EmailClient) setEmailApprovalBody(req core.SubmitRequest, tokenReq string) string {
	var body string
	bodyHeader := `Bismillah
Assalamu'alaikum warahmatullahi wabarakatuh

Hello Admin!
Please review the following job vacancy:
`
	bodyDetailContent := e.setEmailApprovalDetailContent(req)
	bodyApprovalContent := e.setEmailApprovalContent(tokenReq)

	body = fmt.Sprintf("%s\r\n%s\r\n%s", bodyHeader, bodyDetailContent, bodyApprovalContent)

	return body
}

func (e *EmailClient) setEmailApprovalDetailContent(req core.SubmitRequest) string {
	var content string

	if req.SubmissionType == core.SubmitTypeURL {
		content = fmt.Sprintf(`
Submission Email: %s
Submission Type: URL
URL: %s
`, req.SubmissionEmail, req.Vacancy.ApplyURL)
	}

	if req.SubmissionType == core.SubmitTypeManual {
		content = fmt.Sprintf(`
Submission Email: %s
Submission Type: Manual
Job Title: %s
Company: %s
Location: %s
Description: %s
--------------------
Relevant Tags: %s
URL: %s
`,
			req.SubmissionEmail,
			req.Vacancy.JobTitle,
			req.Vacancy.CompanyName,
			req.Vacancy.CompanyLocation,
			req.Vacancy.ShortDescription,
			strings.Join(req.Vacancy.RelevantTags, ", "),
			req.Vacancy.ApplyURL,
		)
	}

	return content
}

func (e *EmailClient) setEmailApprovalContent(tokenReq string) string {
	var content string
	approveLink := fmt.Sprintf("%s/vacancies/approve?data=%s", e.ServerDomain, tokenReq)
	content = fmt.Sprintf(`
To approve, please click the link below:

%s

Thank you and have a nice day! May Allah bless you.

Wassalamu'alaikum warahmatullahi wabarakatuh
`,
		approveLink,
	)

	return content
}

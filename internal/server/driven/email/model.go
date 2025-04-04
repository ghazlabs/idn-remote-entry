package email

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type contentFormat string

const (
	formatHTML        contentFormat = "html"
	formatPlain       contentFormat = "plain"
	generatedIDLength               = 10
)

var possibleChars = []byte("abcdefghijklmnopqrstuvwxyz0123456789")

func generateMessageID(suffix string) string {
	b := make([]byte, generatedIDLength)
	rand.Read(b)
	for i, v := range b {
		b[i] = possibleChars[v%byte(len(possibleChars))]
	}
	return fmt.Sprintf("<%s@%s>", string(b), suffix)
}

func getCodeMessageID(messageID string) string {
	ID := strings.TrimPrefix(messageID, "<")
	ID = strings.Split(ID, "@")[0]
	return ID
}

type templateEmail struct {
	req          core.SubmitRequest
	token        string
	messageID    string
	serverDomain string
}

func (t templateEmail) getContent(format contentFormat) string {
	parts := []string{
		t.getHeader(format),
		t.getDetail(format),
		t.getApproval(format),
	}
	separator := "\r\n"
	if format == formatHTML {
		separator = "<br>\r\n"
	}
	return strings.Join(parts, separator)
}

func (t templateEmail) getContentBodyHTML() string {
	return t.getContent(formatHTML)
}

func (t templateEmail) getContentBodyPlain() string {
	return t.getContent(formatPlain)
}

func (t templateEmail) getHeader(format contentFormat) string {
	if format == formatHTML {
		return `<p>Bismillah<br>Assalamu'alaikum warahmatullahi wabarakatuh</p>
<p>Hello Admin! <br> Please review the following job vacancy:</p>`
	}
	return `Bismillah
Assalamu'alaikum warahmatullahi wabarakatuh

Hello Admin!
Please review the following job vacancy:`
}

func (t templateEmail) getDetail(format contentFormat) string {
	req := t.req
	var lineBreak string
	if format == formatHTML {
		lineBreak = "<br>"
	} else {
		lineBreak = "\n"
	}

	var content strings.Builder

	switch req.SubmissionType {
	case core.SubmitTypeURL:
		fmt.Fprintf(&content, "Submission Email: %s%s", req.SubmissionEmail, lineBreak)
		fmt.Fprintf(&content, "Submission Type: URL%s", lineBreak)
		fmt.Fprintf(&content, "URL: %s", req.Vacancy.ApplyURL)
	case core.SubmitTypeManual:
		fmt.Fprintf(&content, "Submission Email: %s%s", req.SubmissionEmail, lineBreak)
		fmt.Fprintf(&content, "Submission Type: Manual%s", lineBreak)
		fmt.Fprintf(&content, "Job Title: %s%s", req.Vacancy.JobTitle, lineBreak)
		fmt.Fprintf(&content, "Company: %s%s", req.Vacancy.CompanyName, lineBreak)
		fmt.Fprintf(&content, "Location: %s%s", req.Vacancy.CompanyLocation, lineBreak)
		fmt.Fprintf(&content, "Description: %s%s", req.Vacancy.ShortDescription, lineBreak)
		fmt.Fprintf(&content, "Relevant Tags: %s%s", strings.Join(req.Vacancy.RelevantTags, ", "), lineBreak)
		fmt.Fprintf(&content, "URL: %s", req.Vacancy.ApplyURL)
	}

	return content.String()
}

func (t templateEmail) getApproval(format contentFormat) string {
	approveLink := fmt.Sprintf("%s/vacancies/approve?data=%s&message_id=%s", t.serverDomain, t.token, t.messageID)
	rejectLink := fmt.Sprintf("%s/vacancies/reject?data=%s&message_id=%s", t.serverDomain, t.token, t.messageID)

	if format == formatHTML {
		return fmt.Sprintf(`
<div style="margin: 20px 0;">
  <p>Please click one of the buttons below:</p>
  
  <a href="%s" style="display: inline-block; background-color: #4CAF50; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; margin-right: 10px; font-weight: bold;">
  	Approve
  </a>
  
  <a href="%s" style="display: inline-block; background-color: #f44336; color: white; padding: 10px 20px; text-decoration: none; border-radius: 4px; font-weight: bold;">
  	Reject
  </a>
  
  <p style="margin-top: 20px;">
	Thank you and have a nice day!<br>
	Barakallahu fiikum
  </p>
</div>`, approveLink, rejectLink)
	}

	return fmt.Sprintf(`
Please click one of the buttons below:

Approve:

%s

Reject:

%s

Thank you and have a nice day!
Barakallahu fiikum`, approveLink, rejectLink)
}

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
	return fmt.Sprintf("%s@%s", string(b), suffix)
}

func getCodeMessageID(messageID string) string {
	ID := strings.Split(messageID, "@")[0]
	return ID
}

type templateBulkEmail struct {
	req            core.SubmitRequest
	tokenVacancies []string
	messagesIDs    []string
	serverDomain   string
}

func (t templateBulkEmail) getContent(format contentFormat) string {
	parts := []string{
		t.getHeader(format),
		t.getDetail(format),
		t.getVacanciesTable(format),
	}
	separator := "\r\n"
	if format == formatHTML {
		separator = "<br>\r\n"
	}
	return strings.Join(parts, separator)
}

func (t templateBulkEmail) getContentBodyHTML() string {
	return t.getContent(formatHTML)
}

func (t templateBulkEmail) getContentBodyPlain() string {
	return t.getContent(formatPlain)
}

func (t templateBulkEmail) getHeader(format contentFormat) string {
	if format == formatHTML {
		return `<p>Bismillah<br>Assalamu'alaikum warahmatullahi wabarakatuh</p>
<p>Hello Admin! <br> We have found multiple job vacancies that need your review:</p>`
	}
	return `Bismillah
Assalamu'alaikum warahmatullahi wabarakatuh

Hello Admin!
We have found multiple job vacancies that need your review:`
}

func (t templateBulkEmail) getDetail(format contentFormat) string {
	req := t.req
	var lineBreak string
	if format == formatHTML {
		lineBreak = "<br>"
	} else {
		lineBreak = "\n"
	}

	var content strings.Builder
	fmt.Fprintf(&content, "Submission From: %s%s", req.SubmissionEmail, lineBreak)
	fmt.Fprintf(&content, "Submission Type: Bulk%s", lineBreak)
	fmt.Fprintf(&content, "Number of Vacancies: %d%s", len(req.BulkVacancies), lineBreak)

	return content.String()
}

func (t templateBulkEmail) getVacanciesTable(format contentFormat) string {
	if format == formatHTML {
		var tableHTML strings.Builder

		tableHTML.WriteString(`<div style="margin: 20px 0;">
  <table style="width: 100%; border-collapse: collapse; border: 1px solid #ddd;">
    <thead>
      <tr style="background-color: #f2f2f2;">
        <th style="border: 1px solid #ddd; padding: 8px; text-align: left;">No</th>
        <th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Job Title</th>
        <th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Company</th>
        <th style="border: 1px solid #ddd; padding: 8px; text-align: left;">Action</th>
      </tr>
    </thead>
    <tbody>`)

		for i, vacancy := range t.req.BulkVacancies {
			rowStyle := ""
			if i%2 == 1 {
				rowStyle = `style="background-color: #f9f9f9;"`
			}

			approveLink := fmt.Sprintf("%s/vacancies/approve?data=%s&message_id=%s", t.serverDomain, t.tokenVacancies[i], t.messagesIDs[i])
			rejectLink := fmt.Sprintf("%s/vacancies/reject?data=%s&message_id=%s", t.serverDomain, t.tokenVacancies[i], t.messagesIDs[i])

			tableHTML.WriteString(fmt.Sprintf(`
      <tr %s>
        <td style="border: 1px solid #ddd; padding: 8px;">%d</td>
        <td style="border: 1px solid #ddd; padding: 8px;"><a href="%s" target="_blank">%s</a></td>
        <td style="border: 1px solid #ddd; padding: 8px;">%s</td>
        <td style="border: 1px solid #ddd; padding: 8px;">
          <a href="%s" style="display: inline-block; background-color: #4CAF50; color: white; padding: 5px 10px; text-decoration: none; border-radius: 4px; margin-right: 5px; font-size: 12px; font-weight: bold;">
            Approve
          </a>
          <a href="%s" style="display: inline-block; background-color: #f44336; color: white; padding: 5px 10px; text-decoration: none; border-radius: 4px; font-size: 12px; font-weight: bold;">
            Reject
          </a>
        </td>
      </tr>`, rowStyle, i+1, vacancy.ApplyURL, vacancy.JobTitle, vacancy.CompanyName, approveLink, rejectLink))
		}

		tableHTML.WriteString(`
    </tbody>
  </table>
  
  <p style="margin-top: 20px;">
    Thank you for reviewing these vacancies.<br>
    Have a nice day!<br>
    Barakallahu fiikum
  </p>
</div>`)

		return tableHTML.String()
	}

	// Plain text format
	var tablePlain strings.Builder
	tablePlain.WriteString("\nVacancies to review:\n\n")

	for i, vacancy := range t.req.BulkVacancies {
		approveLink := fmt.Sprintf("%s/vacancies/approve?data=%s&message_id=%s", t.serverDomain, t.tokenVacancies[i], t.messagesIDs[i])
		rejectLink := fmt.Sprintf("%s/vacancies/reject?data=%s&message_id=%s", t.serverDomain, t.tokenVacancies[i], t.messagesIDs[i])

		tablePlain.WriteString(fmt.Sprintf("%d. %s at %s\n", i+1, vacancy.JobTitle, vacancy.CompanyName))
		tablePlain.WriteString(fmt.Sprintf("   URL: %s\n", vacancy.ApplyURL))
		tablePlain.WriteString(fmt.Sprintf("   Approve: %s\n", approveLink))
		tablePlain.WriteString(fmt.Sprintf("   Reject: %s\n\n", rejectLink))
	}

	tablePlain.WriteString("\nThank you for reviewing these vacancies.\nHave a nice day!\nBarakallahu fiikum")

	return tablePlain.String()
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
		fmt.Fprintf(&content, "Submission From: %s%s", req.SubmissionEmail, lineBreak)
		fmt.Fprintf(&content, "Submission Type: URL%s", lineBreak)
		fmt.Fprintf(&content, "URL: %s", req.Vacancy.ApplyURL)
	case core.SubmitTypeManual:
		fmt.Fprintf(&content, "Submission From: %s%s", req.SubmissionEmail, lineBreak)
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

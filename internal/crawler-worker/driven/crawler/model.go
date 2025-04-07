package crawler

import (
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

func isEligibleToSave(vacancy core.Vacancy) bool {
	// Check if the vacancy has a valid ApplyURL
	return !strings.Contains(vacancy.ApplyURL, "/cdn-cgi/l/email-protection")
}

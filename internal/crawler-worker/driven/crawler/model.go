package crawler

import (
	"strings"

	"github.com/forPelevin/gomoji"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

func isEligibleToSave(vacancy core.Vacancy) bool {
	// Check if the vacancy has a valid ApplyURL
	return !strings.Contains(vacancy.ApplyURL, "/cdn-cgi/l/email-protection")
}

func removeUnwantedText(text string) string {
	// Remove emojis from the text
	return gomoji.RemoveEmojis(text)
}

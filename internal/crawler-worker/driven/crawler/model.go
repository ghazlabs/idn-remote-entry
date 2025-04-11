package crawler

import (
	"net/url"

	"github.com/forPelevin/gomoji"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

func isEligibleToSave(vacancy core.Vacancy) bool {
	// Parse the URL to validate it
	u, err := url.Parse(vacancy.ApplyURL)
	if err != nil {
		return false
	}

	// Check if URL has https scheme
	return u.Scheme == "https"
}

func removeUnwantedText(text string) string {
	// Remove emojis from the text
	return gomoji.RemoveEmojis(text)
}

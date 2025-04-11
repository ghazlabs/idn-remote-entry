package driver

import (
	"strings"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

func prepareSubmitVacancyRequest(req shcore.SubmitRequest) shcore.SubmitRequest {
	relevantTags := []string{}
	for _, tag := range req.RelevantTags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			relevantTags = append(relevantTags, tag)
		}
	}
	req.RelevantTags = relevantTags

	return req
}

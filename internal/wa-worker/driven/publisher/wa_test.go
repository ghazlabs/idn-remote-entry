package notifier_test

import (
	"context"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/notifier"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/go-resty/resty/v2"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	n, err := notifier.NewWhatsappNotifier(notifier.WhatsappNotifierConfig{
		HttpClient:           resty.New(),
		Username:             env.GetString(testutil.EnvKeyWhatsappApiUser),
		Password:             env.GetString(testutil.EnvKeyWhatsappApiPass),
		WhatsappRecipientIDs: env.GetStrings(testutil.EnvKeyWhatsappRecipientIDs, ","),
	})
	require.NoError(t, err)

	vacRec := core.VacancyRecord{
		ID: "17d35004-6357-8173-ad51-e4a2aa25af5e",
		Vacancy: core.Vacancy{
			JobTitle:         "Happiness Engineer – Customer Support & Success",
			CompanyName:      "Automattic",
			CompanyLocation:  "San Francisco, California, US",
			ShortDescription: "The role involves providing world-class support to customers, helping them use Automattic's products effectively. Responsibilities include troubleshooting, guiding customers, and collaborating with teams to improve user experience. The position requires excellent communication skills, a passion for customer service, and the ability to work independently in a remote environment.",
			RelevantTags:     []string{"customer support", "remote", "communication", "troubleshooting", "wordpress"},
			ApplyURL:         "https://automattic.com/work-with-us/job/happiness-engineer-customer-support-success/",
		},
		PublicURL: "https://idn-remote-jobs.notion.site/Happiness-Engineer-Customer-Support-Success-17b3500463578152adf2fead82be7a4b",
	}

	err = n.Notify(context.Background(), vacRec)
	require.NoError(t, err)
}

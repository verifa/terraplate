package notify

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSlack(t *testing.T) {

	os.Setenv(slackTokenEnv, "TEST")

	sc := DefaultSlackConfig()
	sc.Channel = "test"
	sc.SkipAuthTest = true

	_, notifyErr := New(
		WithNotify(NotifyTypeSlack),
		WithSlackConfig(sc),
	)
	require.NoError(t, notifyErr)
}

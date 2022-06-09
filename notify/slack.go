package notify

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/slack-go/slack"
)

const (
	slackTokenEnv    = "SLACK_API_TOKEN"
	slackChannelEnv  = "SLACK_CHANNEL"
	slackTemplateEnv = "SLACK_TEMPLATE"
	slackSkipAuthEnv = "SLACK_SKIP_AUTH"
)

//go:embed templates/slack.tmpl
var slackTemplate string

var (
	notifyColor          = color.New(color.FgMagenta, color.Bold)
	notifySkipColor      = color.New(color.FgCyan, color.Bold)
	notifySlackMessage   = notifyColor.Sprint("\nSending Slack notification...\n\n")
	notifyNoDriftMessage = notifySkipColor.Sprint("\nNo drift detected. Skipping notification.\n\n")
)

var _ Service = (*slackService)(nil)

func DefaultSlackConfig() *SlackConfig {
	var c SlackConfig
	c.Channel = os.Getenv(slackChannelEnv)
	c.TemplateFile = os.Getenv(slackTemplateEnv)
	_, c.SkipAuthTest = os.LookupEnv(slackSkipAuthEnv)
	return &c
}

type SlackConfig struct {
	// Channel is the slack channel where to send the notification
	Channel string
	// Template is an optional string that will be templated into JSON for
	// Slack's message options.
	// A default template will be used if empty
	TemplateFile string
	// SkipAuthTest defines whether to skip the Slack auth test when creating
	// the service or not
	SkipAuthTest bool
}

func newSlackService(config *SlackConfig, filter NotifyFilter) (Service, error) {
	if config == nil {
		return nil, fmt.Errorf("cannot create Slack service with a nil configuration")
	}
	token, ok := os.LookupEnv(slackTokenEnv)
	if !ok {
		return nil, fmt.Errorf("`%s` env must be set", slackTokenEnv)
	}
	client := slack.New(token)
	if !config.SkipAuthTest {
		_, authErr := client.AuthTest()
		if authErr != nil {
			return nil, fmt.Errorf("auth test failed: %w", authErr)
		}
	}

	// Validate the slack config
	if config.Channel == "" {
		return nil, fmt.Errorf("must provide a slack channel")
	}

	return &slackService{
		Client:       client,
		SlackConfig:  config,
		NotifyFilter: filter,
	}, nil
}

type slackService struct {
	Client       *slack.Client
	SlackConfig  *SlackConfig
	NotifyFilter NotifyFilter
}

type slackMsgJSON struct {
	Attachments []slack.Attachment `json:"attachments,omitempty"`
}

func (s *slackService) Send(data *Data) error {
	if !s.NotifyFilter.ShouldNotify(data.Result) {
		fmt.Print(notifyNoDriftMessage)
		return nil
	}
	fmt.Print(notifySlackMessage)
	msgOpts, optErr := s.messageOptions(data)
	if optErr != nil {
		return fmt.Errorf("creating slack message options: %w", optErr)
	}
	_, _, _, sendErr := s.Client.SendMessage(s.SlackConfig.Channel, msgOpts...)
	if sendErr != nil {
		return fmt.Errorf("sending slack message: %s", sendErr)
	}
	return nil
}

func (s *slackService) messageOptions(data *Data) ([]slack.MsgOption, error) {
	// First we need to execute the template
	var tmplContent = slackTemplate
	if s.SlackConfig.TemplateFile != "" {
		tmplContent = s.SlackConfig.TemplateFile
	}

	tmplBytes, tmplErr := execTemplate(tmplContent, data)
	if tmplErr != nil {
		return nil, fmt.Errorf("templating slack message: %w", tmplErr)
	}

	var msgOpts []slack.MsgOption
	// Set the message text
	now := time.Now()
	msgOpts = append(msgOpts, slack.MsgOptionText(fmt.Sprintf("Terraplate drift on %s", now.Format(time.RFC3339)), false))

	var slackMsg slackMsgJSON
	if err := json.Unmarshal(tmplBytes, &slackMsg); err != nil {
		return nil, fmt.Errorf("unmarshalling slack message: %w", err)
	}
	if slackMsg.Attachments == nil {
		return nil, fmt.Errorf("slack message must have attachments")
	}
	msgOpts = append(msgOpts, slack.MsgOptionAttachments(slackMsg.Attachments...))
	return msgOpts, nil
}

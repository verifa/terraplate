package notify

import (
	"fmt"

	"github.com/verifa/terraplate/runner"
)

type notifyOptsFunc func(n *notifyOpts)

type notifyOpts struct {
	Type        NotifyType
	SlackConfig *SlackConfig

	Service      Service
	NotifyFilter NotifyFilter
}

func WithNotify(notify NotifyType) notifyOptsFunc {
	return func(n *notifyOpts) {
		n.Type = notify
	}
}

func WithSlackConfig(config *SlackConfig) notifyOptsFunc {
	return func(n *notifyOpts) {
		n.SlackConfig = config
	}
}

func NotifyOn(filter NotifyFilter) notifyOptsFunc {
	return func(n *notifyOpts) {
		n.NotifyFilter = filter
	}
}

func New(opts ...notifyOptsFunc) (Service, error) {
	notifier := notifyOpts{
		NotifyFilter: NotifyFilterAll,
	}
	for _, opt := range opts {
		opt(&notifier)
	}

	if !notifier.NotifyFilter.IsValid() {
		return nil, fmt.Errorf("invalid notify filter: %s", notifier.NotifyFilter)
	}

	var (
		service Service
		err     error
	)
	switch notifier.Type {
	case NotifyTypeSlack:
		service, err = newSlackService(notifier.SlackConfig, notifier.NotifyFilter)
	default:
		return nil, fmt.Errorf("unknown notification service type: %s", notifier.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("creating service type %s: %w", notifier.Type, err)
	}
	return service, nil
}

type Service interface {
	Send(data *Data) error
}

type NotifyType string

const (
	NotifyTypeSlack NotifyType = "slack"
)

type NotifyFilter string

const (
	NotifyFilterAll   NotifyFilter = "all"
	NotifyFilterDrift NotifyFilter = "drift"
)

// IsValid checks that the given string for a notification is valid
func (f NotifyFilter) IsValid() bool {
	switch f {
	case NotifyFilterAll, NotifyFilterDrift:
		return true
	}
	return false
}

// ShouldNotify takes a runner and returns a bool whether the notification should
// be sent, or not
func (f NotifyFilter) ShouldNotify(runner *runner.Runner) bool {
	switch f {
	case NotifyFilterAll:
		return true
	case NotifyFilterDrift:
		if runner.HasError() || runner.HasDrift() {
			return true
		}
	}
	return false
}

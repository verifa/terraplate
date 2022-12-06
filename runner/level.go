package runner

import "fmt"

type OutputLevel string

const (
	OutputLevelAll   OutputLevel = "all"
	OutputLevelDrift OutputLevel = "drift"
)

func (o OutputLevel) ShowAll() bool {
	return o == OutputLevelAll
}

func (o OutputLevel) ShowDrift() bool {
	return o == OutputLevelAll || o == OutputLevelDrift
}

func (o OutputLevel) ShowError() bool {
	return true
}

func (o OutputLevel) Validate() (OutputLevel, error) {
	switch o {
	case OutputLevelAll, OutputLevelDrift:
		return o, nil
	default:
		return o, fmt.Errorf("unsupported output level: \"%s\"", o)
	}
}

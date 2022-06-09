package runner

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

type Drift struct {
	AddResources     []*tfjson.ResourceChange
	ChangeResources  []*tfjson.ResourceChange
	DestroyResources []*tfjson.ResourceChange
}

func (d *Drift) HasDrift() bool {
	if len(d.AddResources) == 0 && len(d.ChangeResources) == 0 && len(d.DestroyResources) == 0 {
		return false
	}

	return true
}

func (d *Drift) Diff() string {
	if !d.HasDrift() {
		return planNoChangesColor.Sprint("No changes.")
	}
	return fmt.Sprintf(
		"%s %s %s",
		planCreateColor.Sprintf("%d to add.", len(d.AddResources)),
		planUpdateColor.Sprintf("%d to change.", len(d.ChangeResources)),
		planDestroyColor.Sprintf("%d to destroy.", len(d.DestroyResources)),
	)
}

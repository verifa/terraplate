package runner

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

func driftFromPlan(plan *tfjson.Plan) *Drift {
	var drift Drift
	for _, change := range plan.ResourceChanges {
		for _, action := range change.Change.Actions {
			switch action {
			case tfjson.ActionCreate:
				drift.AddResources = append(drift.AddResources, change)
			case tfjson.ActionDelete:
				drift.DestroyResources = append(drift.DestroyResources, change)
			case tfjson.ActionUpdate:
				drift.ChangeResources = append(drift.ChangeResources, change)
			default:
				// We don't care about other actions for the summary
			}

		}
	}

	return &drift
}

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

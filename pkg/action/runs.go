package action

import (
	"github.com/mikolajgasior/octo-linter/v2/pkg/step"
)

// Runs represents a 'runs' field in a GitHub Actions action parsed from YAML.
type Runs struct {
	Using string       `yaml:"using"`
	Steps []*step.Step `yaml:"steps"`
}

// SetParentType sets type of the parent for all of the steps.
func (ar *Runs) SetParentType(t string) {
	for _, s := range ar.Steps {
		s.ParentType = t
	}
}

// GetStep returns a Step by its id.
func (ar *Runs) GetStep(id string) *step.Step {
	for _, s := range ar.Steps {
		if s.ID == id {
			return s
		}
	}

	return nil
}

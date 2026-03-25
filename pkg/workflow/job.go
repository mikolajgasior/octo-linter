package workflow

import (
	"github.com/mikolajgasior/octo-linter/v2/pkg/step"
)

// Job represents a job in a GitHub Actions workflow parsed from YAML.
type Job struct {
	Name   string            `yaml:"name"`
	Uses   string            `yaml:"uses"`
	RunsOn interface{}       `yaml:"runs-on"`
	Steps  []*step.Step      `yaml:"steps"`
	Env    map[string]string `yaml:"env"`
	Needs  interface{}       `yaml:"needs,omitempty"`
}

// SetParentType sets parent type for all the steps.
func (wj *Job) SetParentType(t string) {
	for _, s := range wj.Steps {
		s.ParentType = t
	}
}

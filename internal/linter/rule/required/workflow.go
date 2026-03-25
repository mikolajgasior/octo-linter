package required

import (
	"fmt"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// Workflow checks if required fields within workflow are defined.
type Workflow struct {
	Field int
}

const (
	_ = iota
	// WorkflowFieldWorkflow specifies that the rule targets top-level fields in a GitHub Actions workflow.
	WorkflowFieldWorkflow
	// WorkflowFieldDispatchInput specifies that the rule targets the 'inputs' section of the 'workflow_dispatch' trigger.
	WorkflowFieldDispatchInput
	// WorkflowFieldCallInput specifies that the rule targets the 'inputs' section of the 'workflow_call' trigger.
	WorkflowFieldCallInput
)

// ConfigName returns the name of the rule as defined in the configuration file.
func (r Workflow) ConfigName(int) string {
	switch r.Field {
	case WorkflowFieldWorkflow:
		return "required_fields__workflow_requires"
	case WorkflowFieldDispatchInput:
		return "required_fields__workflow_dispatch_input_requires"
	case WorkflowFieldCallInput:
		return "required_fields__workflow_call_input_requires"
	default:
		return "required_fields__workflown_*_requires"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r Workflow) FileType() int {
	return rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r Workflow) Validate(conf interface{}) error {
	vals, ok := conf.([]interface{})
	if !ok {
		return errValueNotStringArray
	}

	for _, v := range vals {
		field, ok := v.(string)
		if !ok {
			return errValueNotStringArray
		}

		switch r.Field {
		case WorkflowFieldWorkflow:
			if field != ValueName {
				return errValueNotName
			}
		case WorkflowFieldDispatchInput, WorkflowFieldCallInput:
			if field != ValueDesc {
				return errValueNotDescription
			}
		}
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
//
//nolint:gocognit,funlen
func (r Workflow) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	err := r.Validate(conf)
	if err != nil {
		return false, err
	}

	if file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	workflowInstance, fileIsWorkflow := file.(*workflow.Workflow)
	if !fileIsWorkflow {
		return false, errFileInvalidType
	}

	compliant := true

	confInterfaces, ok := conf.([]interface{})
	if !ok {
		return false, errValueNotStringArray
	}

	switch r.Field {
	case WorkflowFieldWorkflow:
		for _, fieldInterface := range confInterfaces {
			field, ok := fieldInterface.(string)
			if !ok {
				return false, errValueNotStringArray
			}

			if field == ValueName && workflowInstance.Name == "" {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  "does not have a required " + field,
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}

	case WorkflowFieldDispatchInput:
		if workflowInstance.On == nil ||
			workflowInstance.On.WorkflowDispatch == nil ||
			len(workflowInstance.On.WorkflowDispatch.Inputs) == 0 {
			return true, nil
		}

		for inputName, input := range workflowInstance.On.WorkflowDispatch.Inputs {
			for _, fieldInterface := range confInterfaces {
				field, ok := fieldInterface.(string)
				if !ok {
					return false, errValueNotStringArray
				}

				if field == ValueDesc && input.Description == "" {
					chErrors <- glitch.Glitch{
						Path:     workflowInstance.Path,
						Name:     workflowInstance.DisplayName,
						Type:     rule.DotGithubFileTypeWorkflow,
						ErrText:  fmt.Sprintf("dispatch input '%s' does not have a required %s", inputName, field),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	case WorkflowFieldCallInput:
		if workflowInstance.On == nil ||
			workflowInstance.On.WorkflowCall == nil ||
			len(workflowInstance.On.WorkflowCall.Inputs) == 0 {
			return true, nil
		}

		for inputName, input := range workflowInstance.On.WorkflowCall.Inputs {
			for _, fieldInterface := range confInterfaces {
				field, ok := fieldInterface.(string)
				if !ok {
					return false, errValueNotStringArray
				}

				if field == ValueDesc && input.Description == "" {
					chErrors <- glitch.Glitch{
						Path:     workflowInstance.Path,
						Name:     workflowInstance.DisplayName,
						Type:     rule.DotGithubFileTypeWorkflow,
						ErrText:  fmt.Sprintf("call input '%s' does not have a required %s", inputName, field),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	}

	return compliant, nil
}

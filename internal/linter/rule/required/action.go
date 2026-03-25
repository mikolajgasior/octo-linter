package required

import (
	"fmt"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

// Action checks if required fields within actions are defined.
type Action struct {
	Field int
}

const (
	_ = iota
	// ActionFieldAction specifies that the rule targets top-level fields in a GitHub Actions action.
	ActionFieldAction
	// ActionFieldInput specifies that the rule targets the 'inputs' section.
	ActionFieldInput
	// ActionFieldOutput specifies that the rule targets the 'outputs' section.
	ActionFieldOutput
)

const (
	// ValueName defines configuration value for 'name' field.
	ValueName = "name"
	// ValueDesc defines configuration value for 'desc' field.
	ValueDesc = "description"
)

// ConfigName returns the name of the rule as defined in the configuration file.
func (r Action) ConfigName(int) string {
	switch r.Field {
	case ActionFieldAction:
		return "required_fields__action_requires"
	case ActionFieldInput:
		return "required_fields__action_input_requires"
	case ActionFieldOutput:
		return "required_fields__action_output_requires"
	default:
		return "required_fields__action_*_requires"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r Action) FileType() int {
	return rule.DotGithubFileTypeAction
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r Action) Validate(conf interface{}) error {
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
		case ActionFieldAction:
			if field != ValueName && field != ValueDesc {
				return errValueNotNameOrDescription
			}
		case ActionFieldInput, ActionFieldOutput:
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
func (r Action) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	if file.GetType() != rule.DotGithubFileTypeAction {
		return true, nil
	}

	actionInstance, ok := file.(*action.Action)
	if !ok {
		return false, errFileInvalidType
	}

	confInterfaces, confIsInterfaceArray := conf.([]interface{})
	if !confIsInterfaceArray {
		return false, errValueNotStringArray
	}

	compliant := true

	switch r.Field {
	case ActionFieldAction:
		for _, fieldInterface := range confInterfaces {
			field, ok := fieldInterface.(string)
			if !ok {
				return false, errValueNotStringArray
			}

			if (field == ValueName && actionInstance.Name == "") ||
				(field == ValueDesc && actionInstance.Description == "") {
				chErrors <- glitch.Glitch{
					Path:     actionInstance.Path,
					Name:     actionInstance.DirName,
					Type:     rule.DotGithubFileTypeAction,
					ErrText:  "does not have a required " + field,
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	case ActionFieldInput:
		for inputName, input := range actionInstance.Inputs {
			for _, fieldInterface := range confInterfaces {
				field, ok := fieldInterface.(string)
				if !ok {
					return false, errValueNotStringArray
				}

				if field == ValueDesc && input.Description == "" {
					chErrors <- glitch.Glitch{
						Path:     actionInstance.Path,
						Name:     actionInstance.DirName,
						Type:     rule.DotGithubFileTypeAction,
						ErrText:  fmt.Sprintf("input '%s' does not have a required %s", inputName, field),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	case ActionFieldOutput:
		for outputName, output := range actionInstance.Outputs {
			for _, fieldInterface := range confInterfaces {
				field, ok := fieldInterface.(string)
				if !ok {
					return false, errValueNotStringArray
				}

				if field == ValueDesc && output.Description == "" {
					chErrors <- glitch.Glitch{
						Path:     actionInstance.Path,
						Name:     actionInstance.DirName,
						Type:     rule.DotGithubFileTypeAction,
						ErrText:  fmt.Sprintf("output '%s' does not have a required %s", outputName, field),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	}

	return compliant, nil
}

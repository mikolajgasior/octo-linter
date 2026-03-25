package naming

import (
	"fmt"
	"regexp"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/casematch"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

// Action checks if specified action field adheres to the selected naming convention.
type Action struct {
	Field int
}

const (
	_ = iota
	// ActionFieldInputName specifies that the rule targets the action's input name.
	ActionFieldInputName
	// ActionFieldOutputName specifies that the rule targets the action's output name.
	ActionFieldOutputName
	// ActionFieldReferencedVariable specifies that the rule targets all the variables referenced in the action.
	ActionFieldReferencedVariable
	// ActionFieldStepEnv specifies that the rule targets the 'env' section in the action steps.
	ActionFieldStepEnv
)

// ConfigName returns the name of the rule as defined in the configuration file.
func (r Action) ConfigName(int) string {
	switch r.Field {
	case ActionFieldInputName:
		return "naming_conventions__action_input_name_format"
	case ActionFieldOutputName:
		return "naming_conventions__action_output_name_format"
	case ActionFieldReferencedVariable:
		return "naming_conventions__action_referenced_variable_format"
	case ActionFieldStepEnv:
		return "naming_conventions__action_step_env_format"
	default:
		return "naming_conventions__action_*"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r Action) FileType() int {
	return rule.DotGithubFileTypeAction
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r Action) Validate(conf interface{}) error {
	val, ok := conf.(string)
	if !ok {
		return errValueNotString
	}

	if val != "dash-case" && val != "camelCase" && val != "PascalCase" && val != "ALL_CAPS" {
		return errValueNotValid
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
	confValue, confIsString := conf.(string)
	if !confIsString {
		return false, errValueNotString
	}

	if file.GetType() != rule.DotGithubFileTypeAction {
		return true, nil
	}

	actionInstance, ok := file.(*action.Action)
	if !ok {
		return false, errFileInvalidType
	}

	compliant := true

	switch r.Field {
	case ActionFieldInputName:
		for inputName := range actionInstance.Inputs {
			m := casematch.Match(inputName, confValue)
			if !m {
				chErrors <- glitch.Glitch{
					Path:     actionInstance.Path,
					Name:     actionInstance.DirName,
					Type:     rule.DotGithubFileTypeAction,
					ErrText:  fmt.Sprintf("input '%s' must be %s", inputName, confValue),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	case ActionFieldOutputName:
		for outputName := range actionInstance.Outputs {
			m := casematch.Match(outputName, confValue)
			if !m {
				chErrors <- glitch.Glitch{
					Path:     actionInstance.Path,
					Name:     actionInstance.DirName,
					Type:     rule.DotGithubFileTypeAction,
					ErrText:  fmt.Sprintf("output '%s' must be %s", outputName, confValue),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	case ActionFieldReferencedVariable:
		varTypes := []string{"env", "var", "secret"}
		for _, v := range varTypes {
			re := regexp.MustCompile(fmt.Sprintf("\\${{[ ]*%s\\.([a-zA-Z0-9\\-_]+)[ ]*}}", v))

			found := re.FindAllSubmatch(actionInstance.Raw, -1)
			for _, refVar := range found {
				m := casematch.Match(string(refVar[1]), confValue)
				if !m {
					chErrors <- glitch.Glitch{
						Path:     actionInstance.Path,
						Name:     actionInstance.DirName,
						Type:     rule.DotGithubFileTypeAction,
						ErrText:  fmt.Sprintf("references a variable '%s' that must be %s", string(refVar[1]), confValue),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	case ActionFieldStepEnv:
		if len(actionInstance.Runs.Steps) == 0 {
			return true, nil
		}

		for stepIdx, step := range actionInstance.Runs.Steps {
			if len(step.Env) == 0 {
				continue
			}

			for envName := range step.Env {
				m := casematch.Match(envName, confValue)
				if !m {
					chErrors <- glitch.Glitch{
						Path:     actionInstance.Path,
						Name:     actionInstance.DirName,
						Type:     rule.DotGithubFileTypeAction,
						ErrText:  fmt.Sprintf("step %d env '%s' must be %s", stepIdx, envName, confValue),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	}

	return compliant, nil
}

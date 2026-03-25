package refvars

import (
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// NotOneWord checks for variable references that are single-word or single-level, e.g. `${{ something }}` instead of
// `${{ inputs.something }}`.
// Only the values `true` and `false` are permitted in this form; all other variables are considered invalid.
type NotOneWord struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r NotOneWord) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "referenced_variables_in_workflows__not_one_word"
	case rule.DotGithubFileTypeAction:
		return "referenced_variables_in_actions__not_one_word"
	default:
		return "referenced_variables_in_*__not_one_word"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r NotOneWord) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r NotOneWord) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r NotOneWord) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsBool := conf.(bool)
	if !confIsBool {
		return false, errValueNotBool
	}

	if file.GetType() != rule.DotGithubFileTypeAction &&
		file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	if !confValue {
		return true, nil
	}

	compliant := true

	if file.GetType() == rule.DotGithubFileTypeAction {
		actionInstance, ok := file.(*action.Action)
		if !ok {
			return false, errFileInvalidType
		}

		foundNotCompliant := processActionForRegexp(
			r.ConfigName(rule.DotGithubFileTypeAction),
			actionInstance,
			regexpReference,
			chErrors,
			"invalid",
		)
		if foundNotCompliant {
			compliant = false
		}
	}

	if file.GetType() == rule.DotGithubFileTypeWorkflow {
		workflowInstance, ok := file.(*workflow.Workflow)
		if !ok {
			return false, errFileInvalidType
		}

		foundNotCompliant := processWorkflowForRegexp(
			r.ConfigName(rule.DotGithubFileTypeWorkflow),
			workflowInstance,
			regexpReference,
			chErrors,
			"invalid",
		)
		if foundNotCompliant {
			compliant = false
		}
	}

	return compliant, nil
}

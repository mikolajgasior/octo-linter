package refvars

import (
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// NotInDoubleQuotes scans for all variable references enclosed in double quotes. It is safer to use single quotes, as
// double quotes expand certain characters and may allow the execution of sub-commands.
type NotInDoubleQuotes struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r NotInDoubleQuotes) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "referenced_variables_in_workflows__not_in_double_quotes"
	case rule.DotGithubFileTypeAction:
		return "referenced_variables_in_actions__not_in_double_quotes"
	default:
		return "referenced_variables_in_*__not_in_double_quotes"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r NotInDoubleQuotes) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r NotInDoubleQuotes) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r NotInDoubleQuotes) Lint(
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
			regexpReferenceInDoubleQuote,
			chErrors,
			"in double quotes",
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
			regexpReferenceInDoubleQuote,
			chErrors,
			"in double quotes",
		)
		if foundNotCompliant {
			compliant = false
		}
	}

	return compliant, nil
}

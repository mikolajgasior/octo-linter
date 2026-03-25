package filenames

import (
	"strings"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/casematch"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// WorkflowFilenameBaseFormat checks if workflow file basename (without extension) adheres to the selected naming
// convention.
type WorkflowFilenameBaseFormat struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r WorkflowFilenameBaseFormat) ConfigName(int) string {
	return "filenames__workflow_filename_base_format"
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r WorkflowFilenameBaseFormat) FileType() int {
	return rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r WorkflowFilenameBaseFormat) Validate(conf interface{}) error {
	val, ok := conf.(string)
	if !ok {
		return errValueNotString
	}

	if val != ValueDashCase && val != ValueDashCaseUnderscore && val != ValueCamelCase &&
		val != ValuePascalCase &&
		val != ValueAllCaps {
		return errValueNotValidIncludingDashCaseUnderscore
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r WorkflowFilenameBaseFormat) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsString := conf.(string)
	if !confIsString {
		return false, errValueNotString
	}

	if file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	workflowInstance, ok := file.(*workflow.Workflow)
	if !ok {
		return false, errFileInvalidType
	}

	fileParts := strings.Split(workflowInstance.FileName, ".")
	basename := fileParts[0]

	m := casematch.Match(basename, confValue)
	if !m {
		chErrors <- glitch.Glitch{
			Path:     workflowInstance.Path,
			Name:     workflowInstance.DisplayName,
			Type:     rule.DotGithubFileTypeWorkflow,
			ErrText:  "filename base must be " + confValue,
			RuleName: r.ConfigName(0),
		}

		return false, nil
	}

	return true, nil
}

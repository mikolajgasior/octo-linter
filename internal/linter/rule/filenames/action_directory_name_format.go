package filenames

import (
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/casematch"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

// ActionDirectoryNameFormat checks if the directory containing the action adheres to the selected naming convention.
type ActionDirectoryNameFormat struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r ActionDirectoryNameFormat) ConfigName(int) string {
	return "filenames__action_directory_name_format"
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r ActionDirectoryNameFormat) FileType() int {
	return rule.DotGithubFileTypeAction
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r ActionDirectoryNameFormat) Validate(conf interface{}) error {
	val, ok := conf.(string)
	if !ok {
		return errValueNotString
	}

	if val != ValueDashCase && val != ValueCamelCase && val != ValuePascalCase &&
		val != ValueAllCaps {
		return errValueNotValid
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r ActionDirectoryNameFormat) Lint(
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

	m := casematch.Match(actionInstance.DirName, confValue)
	if !m {
		chErrors <- glitch.Glitch{
			Path:     actionInstance.Path,
			Name:     actionInstance.DirName,
			Type:     rule.DotGithubFileTypeAction,
			ErrText:  "directory name must be " + confValue,
			RuleName: r.ConfigName(0),
		}

		return false, nil
	}

	return true, nil
}

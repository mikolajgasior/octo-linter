package usedactions

import (
	"fmt"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/step"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// Source checks if referenced action (in `uses`) in steps has valid path.
// This rule can be configured to allow local actions, external actions, or both.
type Source struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r Source) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "used_actions_in_workflow_job_steps__source"
	case rule.DotGithubFileTypeAction:
		return "used_actions_in_action_steps__source"
	default:
		return "used_actions_in_*_steps__source"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r Source) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r Source) Validate(conf interface{}) error {
	val, ok := conf.(string)
	if !ok {
		return errValueNotString
	}

	if val != ValueLocalOnly && val != ValueLocalOrExternal && val != ValueExternalOnly &&
		val != "" {
		return errValueNotEmptyOrLocalOrExternalOrBoth
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r Source) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsString := conf.(string)
	if !confIsString {
		return false, errValueNotString
	}

	if file.GetType() != rule.DotGithubFileTypeAction &&
		file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	if confValue == "" {
		return true, nil
	}

	var (
		steps     []*step.Step
		msgPrefix map[int]string
		fileType  int
		filePath  string
		fileName  string
	)

	if file.GetType() == rule.DotGithubFileTypeAction {
		actionInstance, ok := file.(*action.Action)
		if !ok {
			return false, errFileInvalidType
		}

		if len(actionInstance.Runs.Steps) == 0 {
			return true, nil
		}

		steps, msgPrefix, fileType, filePath, fileName = getStepsFromAction(actionInstance)
	}

	if file.GetType() == rule.DotGithubFileTypeWorkflow {
		workflowInstance, ok := file.(*workflow.Workflow)
		if !ok {
			return false, errFileInvalidType
		}

		if len(workflowInstance.Jobs) == 0 {
			return true, nil
		}

		steps, msgPrefix, fileType, filePath, fileName = getStepsFromWorkflow(workflowInstance)
	}

	compliant := r.processSteps(steps, msgPrefix, fileType, filePath, fileName, chErrors, confValue)

	return compliant, nil
}

//nolint:funlen
func (r Source) processSteps(
	steps []*step.Step,
	msgPrefix map[int]string,
	fileType int,
	filePath string,
	fileName string,
	chErrors chan<- glitch.Glitch,
	confValue string,
) bool {
	var errPrefix string
	if fileType == rule.DotGithubFileTypeAction {
		errPrefix = msgPrefix[0]
	}

	compliant := true

	for stepIdx, step := range steps {
		newErrPrefix, ok := msgPrefix[stepIdx]
		if ok {
			errPrefix = newErrPrefix
		}

		if step.Uses == "" {
			continue
		}

		isLocal := regexpLocalAction.MatchString(step.Uses)
		isExternal := regexpExternalAction.MatchString(step.Uses)

		if confValue == ValueLocalOnly && !isLocal {
			chErrors <- glitch.Glitch{
				Path: filePath,
				Name: fileName,
				Type: fileType,
				ErrText: fmt.Sprintf(
					"%sstep %d calls action '%s' that is not a valid local path",
					errPrefix,
					stepIdx+1,
					step.Uses,
				),
				RuleName: r.ConfigName(fileType),
			}

			compliant = false
		}

		if confValue == ValueExternalOnly && !isExternal {
			chErrors <- glitch.Glitch{
				Path: filePath,
				Name: fileName,
				Type: fileType,
				ErrText: fmt.Sprintf(
					"%sstep %d calls action '%s' that is not external",
					errPrefix,
					stepIdx+1,
					step.Uses,
				),
				RuleName: r.ConfigName(fileType),
			}

			compliant = false
		}

		if confValue == ValueLocalOrExternal && !isLocal && !isExternal {
			chErrors <- glitch.Glitch{
				Path: filePath,
				Name: fileName,
				Type: fileType,
				ErrText: fmt.Sprintf(
					"%sstep %d calls action '%s' that is neither external nor local",
					errPrefix,
					stepIdx+1,
					step.Uses,
				),
				RuleName: r.ConfigName(fileType),
			}

			compliant = false
		}
	}

	return compliant
}

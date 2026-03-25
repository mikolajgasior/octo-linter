package usedactions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/step"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// Exists verifies that the action referenced in a step actually exists.
type Exists struct{}

var errConfValue = errors.New("config value is invalid")

func errGettingConfValue(err error) error {
	return fmt.Errorf("%w: %s", errConfValue, err.Error())
}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r Exists) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "used_actions_in_workflow_job_steps__must_exist"
	case rule.DotGithubFileTypeAction:
		return "used_actions_in_action_steps__must_exist"
	default:
		return "used_actions_in_*_steps__must_exist"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r Exists) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r Exists) Validate(conf interface{}) error {
	vals, ok := conf.([]interface{})
	if !ok {
		return errValueNotStringArray
	}

	for _, v := range vals {
		source, ok := v.(string)
		if !ok {
			return errValueNotStringArray
		}

		if source != "local" && source != "external" {
			return errValueNotLocalAndOrExternal
		}
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r Exists) Lint(
	conf interface{},
	file dotgithub.File,
	dotGithub *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	if file.GetType() != rule.DotGithubFileTypeAction &&
		file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	checkLocal, checkExternal, err := r.getChecks(conf)
	if err != nil {
		return false, errGettingConfValue(err)
	}

	if !checkLocal && !checkExternal {
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

	compliant := r.processSteps(
		steps,
		msgPrefix,
		fileType,
		filePath,
		fileName,
		chErrors,
		checkLocal,
		checkExternal,
		dotGithub,
	)

	return compliant, nil
}

func (r Exists) processSteps(
	steps []*step.Step,
	msgPrefix map[int]string,
	fileType int,
	filePath string,
	fileName string,
	chErrors chan<- glitch.Glitch,
	checkLocal bool,
	checkExternal bool,
	dotGithub *dotgithub.DotGithub,
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

		if checkLocal && isLocal {
			actionName := strings.ReplaceAll(step.Uses, "./.github/actions/", "")

			action := dotGithub.GetAction(actionName)
			if action == nil {
				compliant = false

				chErrors <- glitch.Glitch{
					Path:     filePath,
					Name:     fileName,
					Type:     fileType,
					ErrText:  fmt.Sprintf("%sstep %d calls non-existing local action '%s'", errPrefix, stepIdx+1, actionName),
					RuleName: r.ConfigName(fileType),
				}
			}
		}

		if checkExternal && isExternal {
			action := dotGithub.GetExternalAction(step.Uses)
			if action == nil {
				compliant = false

				chErrors <- glitch.Glitch{
					Path:     filePath,
					Name:     fileName,
					Type:     fileType,
					ErrText:  fmt.Sprintf("%sstep %d calls non-existing external action '%s'", errPrefix, stepIdx+1, step.Uses),
					RuleName: r.ConfigName(fileType),
				}
			}
		}
	}

	return compliant
}

func (r Exists) getChecks(conf interface{}) (bool, bool, error) {
	var (
		checkLocal    bool
		checkExternal bool
	)

	valInterfaces, confIsInterfaceArray := conf.([]interface{})
	if !confIsInterfaceArray {
		return false, false, errValueNotStringArray
	}

	for _, valInterface := range valInterfaces {
		val, ok := valInterface.(string)
		if !ok {
			return false, false, errValueNotStringArray
		}

		if val == "local" {
			checkLocal = true
		}

		if val == "external" {
			checkExternal = true
		}
	}

	return checkLocal, checkExternal, nil
}

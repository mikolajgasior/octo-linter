package usedactions

import (
	"fmt"
	"strings"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/step"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// ValidInputs verifies that all required inputs are provided when referencing an action in a step, and that no
// undefined inputs are used.
type ValidInputs struct {
	FileTypeRequired string
}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r ValidInputs) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "used_actions_in_workflow_job_steps__must_have_valid_inputs"
	case rule.DotGithubFileTypeAction:
		return "used_actions_in_action_steps__must_have_valid_inputs"
	default:
		return "used_actions_in_*_steps__must_have_valid_inputs"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r ValidInputs) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r ValidInputs) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r ValidInputs) Lint(
	conf interface{},
	file dotgithub.File,
	dotGithub *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsBool := conf.(bool)
	if !confIsBool {
		return false, errValueNotBool
	}

	var fileTypeRequired int
	if r.FileTypeRequired == "action" {
		fileTypeRequired = rule.DotGithubFileTypeAction
	}
	if r.FileTypeRequired == "workflow" {
		fileTypeRequired = rule.DotGithubFileTypeWorkflow
	}

	if file.GetType() != fileTypeRequired {
		return true, nil
	}

	if !confValue {
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

	compliant := r.processSteps(steps, msgPrefix, fileType, filePath, fileName, chErrors, dotGithub)

	return compliant, nil
}

func (r ValidInputs) processStepActionInputs(
	stepActionInputs map[string]*action.Input,
	step *step.Step,
	stepIdx int,
	errPrefix string,
	chErrors chan<- glitch.Glitch,
	fileType int,
	filePath string,
	fileName string,
) bool {
	if stepActionInputs == nil {
		return false
	}

	foundNotCompliant := false

	for daInputName, daInput := range stepActionInputs {
		if !daInput.Required {
			continue
		}

		if step.With != nil && step.With[daInputName] != "" {
			continue
		}

		chErrors <- glitch.Glitch{
			Path:     filePath,
			Name:     fileName,
			Type:     fileType,
			ErrText:  fmt.Sprintf("%sstep %d called action requires input '%s'", errPrefix, stepIdx+1, daInputName),
			RuleName: r.ConfigName(fileType),
		}

		foundNotCompliant = true
	}

	return foundNotCompliant
}

func (r ValidInputs) processStepWith(
	stepWith map[string]string,
	stepActionInputs map[string]*action.Input,
	stepIdx int,
	errPrefix string,
	chErrors chan<- glitch.Glitch,
	fileType int,
	filePath string,
	fileName string,
) bool {
	if stepWith == nil {
		return false
	}

	foundNotCompliant := false

	for usedInput := range stepWith {
		if stepActionInputs != nil && stepActionInputs[usedInput] != nil {
			continue
		}

		chErrors <- glitch.Glitch{
			Path:     filePath,
			Name:     fileName,
			Type:     fileType,
			ErrText:  fmt.Sprintf("%sstep %d called action non-existing input '%s'", errPrefix, stepIdx+1, usedInput),
			RuleName: r.ConfigName(fileType),
		}

		foundNotCompliant = true
	}

	return foundNotCompliant
}

func (r ValidInputs) processSteps(
	steps []*step.Step,
	msgPrefix map[int]string,
	fileType int,
	filePath string,
	fileName string,
	chErrors chan<- glitch.Glitch,
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
		stepAction := r.getStepFromStepUses(step.Uses, dotGithub)
		if stepAction == nil {
			continue
		}

		if r.processStepActionInputs(
			stepAction.Inputs,
			step,
			stepIdx,
			errPrefix,
			chErrors,
			fileType,
			filePath,
			fileName,
		) {
			compliant = false
		}

		if r.processStepWith(
			step.With,
			stepAction.Inputs,
			stepIdx,
			errPrefix,
			chErrors,
			fileType,
			filePath,
			fileName,
		) {
			compliant = false
		}
	}

	return compliant
}

func (r ValidInputs) getStepFromStepUses(
	stepUses string,
	dotGithub *dotgithub.DotGithub,
) *action.Action {
	if stepUses == "" {
		return nil
	}
	isLocal := regexpLocalAction.MatchString(stepUses)
	isExternal := regexpExternalAction.MatchString(stepUses)
	if isLocal {
		actionName := strings.ReplaceAll(stepUses, "./.github/actions/", "")

		return dotGithub.GetAction(actionName)
	} else if isExternal {
		return dotGithub.GetExternalAction(stepUses)
	}

	return nil
}

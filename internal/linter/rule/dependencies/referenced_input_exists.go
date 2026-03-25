package dependencies

import (
	"fmt"
	"regexp"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// ReferencedInputExists scans the code for all input references and verifies that each has been previously defined.
// During action or workflow execution, if a reference to an undefined input is found, it is replaced with an empty
// string.
type ReferencedInputExists struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r ReferencedInputExists) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "dependencies__workflow_referenced_input_must_exists"
	case rule.DotGithubFileTypeAction:
		return "dependencies__action_referenced_input_must_exists"
	default:
		return "dependencies__*_referenced_input_must_exists"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r ReferencedInputExists) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r ReferencedInputExists) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r ReferencedInputExists) Lint(
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

		foundNotCompliant := r.processAction(actionInstance, chErrors)
		if foundNotCompliant {
			compliant = false
		}

		return compliant, nil
	}

	if file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	// check workflow
	workflowInstance, ok := file.(*workflow.Workflow)
	if !ok {
		return false, errFileInvalidType
	}

	foundNotCompliant := r.processWorkflow(workflowInstance, chErrors)
	if foundNotCompliant {
		compliant = false
	}

	return compliant, nil
}

func (r ReferencedInputExists) processAction(
	actionInstance *action.Action,
	chErrors chan<- glitch.Glitch,
) bool {
	foundNotCompliant := false

	re := regexp.MustCompile(regexpRefInput)

	found := re.FindAllSubmatch(actionInstance.Raw, -1)
	for _, refInput := range found {
		if actionInstance.Inputs == nil || actionInstance.Inputs[string(refInput[1])] == nil {
			chErrors <- glitch.Glitch{
				Path:     actionInstance.Path,
				Name:     actionInstance.DirName,
				Type:     rule.DotGithubFileTypeAction,
				ErrText:  fmt.Sprintf("calls an input '%s' that does not exist", string(refInput[1])),
				RuleName: r.ConfigName(rule.DotGithubFileTypeAction),
			}

			foundNotCompliant = true
		}
	}

	return foundNotCompliant
}

func (r ReferencedInputExists) processWorkflow(
	workflowInstance *workflow.Workflow,
	chErrors chan<- glitch.Glitch,
) bool {
	foundNotCompliant := false

	re := regexp.MustCompile(regexpRefInput)

	found := re.FindAllSubmatch(workflowInstance.Raw, -1)
	for _, refInput := range found {
		notInInputs := true

		if workflowInstance.On != nil {
			if workflowInstance.On.WorkflowCall != nil &&
				workflowInstance.On.WorkflowCall.Inputs != nil &&
				workflowInstance.On.WorkflowCall.Inputs[string(refInput[1])] != nil {
				notInInputs = false
			}

			if workflowInstance.On.WorkflowDispatch != nil &&
				workflowInstance.On.WorkflowDispatch.Inputs != nil &&
				workflowInstance.On.WorkflowDispatch.Inputs[string(refInput[1])] != nil {
				notInInputs = false
			}
		}

		if !notInInputs {
			continue
		}

		chErrors <- glitch.Glitch{
			Path:     workflowInstance.Path,
			Name:     workflowInstance.DisplayName,
			Type:     rule.DotGithubFileTypeWorkflow,
			ErrText:  fmt.Sprintf("calls an input '%s' that does not exist", string(refInput[1])),
			RuleName: r.ConfigName(rule.DotGithubFileTypeWorkflow),
		}

		foundNotCompliant = true
	}

	return foundNotCompliant
}

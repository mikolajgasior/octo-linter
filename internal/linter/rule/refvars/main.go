// Package refvars contains rules checking variables referenced in action or workflow steps, eg. ${{ var }}.
package refvars

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

var (
	errFileInvalidType = errors.New("file is of invalid type")
	errValueNotBool    = errors.New("value should be bool")
)

const (
	regexpReferenceInDoubleQuote = `\"\${{[ ]*([a-zA-Z0-9\-_.]+)[ ]*}}\"`
	regexpReference              = `\${{[ ]*([a-zA-Z0-9\-_]+)[ ]*}}`
)

func processActionForRegexp(
	ruleConfigName string,
	actionInstance *action.Action,
	regexpToMatch string,
	chErrors chan<- glitch.Glitch,
	errorText string,
) bool {
	foundNotCompliant := false

	refRegexp := regexp.MustCompile(regexpToMatch)

	found := refRegexp.FindAllSubmatch(actionInstance.Raw, -1)
	for _, ref := range found {
		chErrors <- glitch.Glitch{
			Path:     actionInstance.Path,
			Name:     actionInstance.DirName,
			Type:     rule.DotGithubFileTypeAction,
			ErrText:  fmt.Sprintf("calls a variable '%s' that is %s", string(ref[1]), errorText),
			RuleName: ruleConfigName,
		}

		foundNotCompliant = true
	}

	return foundNotCompliant
}

func processWorkflowForRegexp(
	ruleConfigName string,
	workflowInstance *workflow.Workflow,
	regexpToMatch string,
	chErrors chan<- glitch.Glitch,
	errorText string,
) bool {
	foundNotCompliant := false

	refRegexp := regexp.MustCompile(regexpToMatch)

	found := refRegexp.FindAllSubmatch(workflowInstance.Raw, -1)
	for _, ref := range found {
		chErrors <- glitch.Glitch{
			Path:     workflowInstance.Path,
			Name:     workflowInstance.DisplayName,
			Type:     rule.DotGithubFileTypeWorkflow,
			ErrText:  fmt.Sprintf("calls a variable '%s' that is %s", string(ref[1]), errorText),
			RuleName: ruleConfigName,
		}

		foundNotCompliant = true
	}

	return foundNotCompliant
}

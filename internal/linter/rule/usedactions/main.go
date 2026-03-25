// Package usedactions contains rules checking paths of actions used in steps.
package usedactions

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/step"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

const (
	// ValueLocalOnly defines a configuration value for the referenced action (in 'uses' field) to be local only.
	ValueLocalOnly = "local-only"
	// ValueExternalOnly defines a configuration value for the referenced action (in 'uses' field) to be external only.
	ValueExternalOnly = "external-only"
	// ValueLocalOrExternal defines a configuration value for the referenced action (in 'uses' field) to be local or
	// external.
	ValueLocalOrExternal = "local-or-external"
)

var (
	errValueNotBool               = errors.New("value should be bool")
	errValueNotString             = errors.New("value should be string")
	errValueNotStringArray        = errors.New("value should be []string")
	errValueNotLocalAndOrExternal = errors.New(
		"value can contain only 'local' and/or 'external'",
	)
	errValueNotEmptyOrLocalOrExternalOrBoth = fmt.Errorf(
		"value can be '%s', '%s', '%s' or empty string",
		ValueLocalOnly,
		ValueLocalOrExternal,
		ValueExternalOnly,
	)
	errFileInvalidType = errors.New("file is of invalid type")
)

var (
	regexpLocalAction = regexp.MustCompile(
		`^\.\/\.github\/actions\/([a-zA-Z0-9\-_]+|[a-zA-Z0-9\-\_]+\/[a-zA-Z0-9\-_]+)$`,
	)
	regexpExternalAction = regexp.MustCompile(
		`[a-zA-Z0-9\-\_]+\/[a-zA-Z0-9\-\_]+(\/[a-zA-Z0-9\-\_]){0,1}@[a-zA-Z0-9\.\-\_]+`,
	)
)

func getStepsFromAction(
	actionInstance *action.Action,
) ([]*step.Step, map[int]string, int,
	string, string,
) {
	steps := actionInstance.Runs.Steps

	msgPrefix := map[int]string{0: ""}

	fileType := rule.DotGithubFileTypeAction
	filePath := actionInstance.Path
	fileName := actionInstance.DirName

	return steps, msgPrefix, fileType, filePath, fileName
}

func getStepsFromWorkflow(
	workflowInstance *workflow.Workflow,
) ([]*step.Step, map[int]string,
	int, string, string,
) {
	steps := []*step.Step{}
	msgPrefix := map[int]string{}

	for jobName, job := range workflowInstance.Jobs {
		if len(job.Steps) == 0 {
			continue
		}

		msgPrefix[len(steps)] = fmt.Sprintf("job '%s'", jobName)

		steps = append(steps, job.Steps...)
	}

	fileType := rule.DotGithubFileTypeWorkflow
	filePath := workflowInstance.Path
	fileName := workflowInstance.DisplayName

	return steps, msgPrefix, fileType, filePath, fileName
}

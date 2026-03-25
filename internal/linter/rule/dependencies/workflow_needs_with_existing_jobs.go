package dependencies

import (
	"fmt"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// WorkflowNeedsWithExistingJobs checks if `needs` field references existing jobs.
type WorkflowNeedsWithExistingJobs struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r WorkflowNeedsWithExistingJobs) ConfigName(int) string {
	return "dependencies__workflow_needs_field_must_contain_already_existing_jobs"
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r WorkflowNeedsWithExistingJobs) FileType() int {
	return rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r WorkflowNeedsWithExistingJobs) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r WorkflowNeedsWithExistingJobs) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsBool := conf.(bool)
	if !confIsBool {
		return false, errValueNotBool
	}

	if file.GetType() != rule.DotGithubFileTypeWorkflow || confValue {
		return true, nil
	}

	workflowInstance, ok := file.(*workflow.Workflow)
	if !ok {
		return false, errFileInvalidType
	}

	if len(workflowInstance.Jobs) == 0 {
		return true, nil
	}

	compliant := true

	for jobName, job := range workflowInstance.Jobs {
		if job.Needs == nil {
			continue
		}

		foundNotCompliant := r.processJobNeeds(workflowInstance, jobName, job.Needs, chErrors)
		if foundNotCompliant {
			compliant = false
		}
	}

	return compliant, nil
}

func (r WorkflowNeedsWithExistingJobs) processJobNeeds(
	workflowInstance *workflow.Workflow,
	jobName string,
	jobNeeds interface{},
	chErrors chan<- glitch.Glitch,
) bool {
	needsStr, needsIsString := jobNeeds.(string)
	if needsIsString {
		if workflowInstance.Jobs[needsStr] != nil {
			return false
		}

		chErrors <- glitch.Glitch{
			Path:     workflowInstance.Path,
			Name:     workflowInstance.DisplayName,
			Type:     rule.DotGithubFileTypeWorkflow,
			ErrText:  fmt.Sprintf("job '%s' has non-existing job '%s' in 'needs' field", jobName, needsStr),
			RuleName: r.ConfigName(0),
		}

		return true
	}

	needsList, needsIsList := jobNeeds.([]interface{})
	if !needsIsList {
		return false
	}

	foundNotCompliant := false

	for _, neededJobInterface := range needsList {
		neededJob, ok := neededJobInterface.(string)
		if !ok {
			return false
		}

		if workflowInstance.Jobs[neededJob] != nil {
			return false
		}

		chErrors <- glitch.Glitch{
			Path:     workflowInstance.Path,
			Name:     workflowInstance.DisplayName,
			Type:     rule.DotGithubFileTypeWorkflow,
			ErrText:  fmt.Sprintf("job '%s' has non-existing job '%s' in 'needs' field", jobName, neededJob),
			RuleName: r.ConfigName(0),
		}

		foundNotCompliant = true
	}

	return foundNotCompliant
}

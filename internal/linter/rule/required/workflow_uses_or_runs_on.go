package required

import (
	"fmt"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// WorkflowUsesOrRunsOn checks if workflow has `runs-on` or `uses` field. At least of them must be defined.
type WorkflowUsesOrRunsOn struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r WorkflowUsesOrRunsOn) ConfigName(int) string {
	return "required_fields__workflow_requires_uses_or_runs_on_required"
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r WorkflowUsesOrRunsOn) FileType() int {
	return rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r WorkflowUsesOrRunsOn) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r WorkflowUsesOrRunsOn) Lint(
	conf interface{},
	file dotgithub.File,
	_ *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsBool := conf.(bool)
	if !confIsBool {
		return false, errValueNotBool
	}

	if file.GetType() != rule.DotGithubFileTypeWorkflow {
		return true, nil
	}

	workflowInstance, ok := file.(*workflow.Workflow)
	if !ok {
		return false, errFileInvalidType
	}

	if !confValue || len(workflowInstance.Jobs) == 0 {
		return true, nil
	}

	compliant := true

	for jobName, job := range workflowInstance.Jobs {
		if job.RunsOn == nil && job.Uses == "" {
			chErrors <- glitch.Glitch{
				Path:     workflowInstance.Path,
				Name:     workflowInstance.DisplayName,
				Type:     rule.DotGithubFileTypeWorkflow,
				ErrText:  fmt.Sprintf("job '%s' should have either 'uses' or 'runs-on' field", jobName),
				RuleName: r.ConfigName(0),
			}

			compliant = false
		}

		runsOnStr, ok := job.RunsOn.(string)
		if ok {
			if job.Uses == "" && runsOnStr == "" {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("job '%s' should have either 'uses' or 'runs-on' field", jobName),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	}

	return compliant, nil
}

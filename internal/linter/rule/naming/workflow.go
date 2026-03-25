package naming

import (
	"fmt"
	"regexp"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/casematch"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// Workflow checks if specified workflow field adheres to the selected naming convention.
type Workflow struct {
	Field int
}

const (
	_ = iota
	// WorkflowFieldEnv specifies that the rule targets the top-level 'env' section.
	WorkflowFieldEnv
	// WorkflowFieldJobEnv specifies that the rule targets the 'env' section in jobs.
	WorkflowFieldJobEnv
	// WorkflowFieldJobStepEnv specifies that the rule targets the 'env' section in steps of each job.
	WorkflowFieldJobStepEnv
	// WorkflowFieldReferencedVariable specifies that the rule targets all the variables referenced in the workflow.
	WorkflowFieldReferencedVariable
	// WorkflowFieldDispatchInputName specifies that the rule targets the input names of the 'workflow_dispatch' trigger.
	WorkflowFieldDispatchInputName
	// WorkflowFieldCallInputName specifies that the rule targets the input names of the 'workflow_call' trigger.
	WorkflowFieldCallInputName
	// WorkflowFieldJobName specifies that the rule targets names of the jobs.
	WorkflowFieldJobName
)

// ConfigName returns the name of the rule as defined in the configuration file.
func (r Workflow) ConfigName(int) string {
	switch r.Field {
	case WorkflowFieldEnv:
		return "naming_conventions__workflow_env_format"
	case WorkflowFieldJobEnv:
		return "naming_conventions__workflow_job_env_format"
	case WorkflowFieldJobStepEnv:
		return "naming_conventions__workflow_job_step_env_format"
	case WorkflowFieldReferencedVariable:
		return "naming_conventions__workflow_referenced_variable_format"
	case WorkflowFieldDispatchInputName:
		return "naming_conventions__workflow_dispatch_input_name_format"
	case WorkflowFieldCallInputName:
		return "naming_conventions__workflow_call_input_name_format"
	case WorkflowFieldJobName:
		return "naming_conventions__workflow_job_name_format"
	default:
		return "naming_conventions__workflow_*"
	}
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r Workflow) FileType() int {
	return rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r Workflow) Validate(conf interface{}) error {
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
//
//nolint:gocognit,funlen,gocyclo
func (r Workflow) Lint(
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

	compliant := true

	switch r.Field {
	case WorkflowFieldEnv:
		if len(workflowInstance.Env) == 0 {
			return true, nil
		}

		for envName := range workflowInstance.Env {
			m := casematch.Match(envName, confValue)
			if !m {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("env '%s' must be %s", envName, confValue),
					RuleName: r.ConfigName(0),
				}
			}
		}
	case WorkflowFieldJobEnv:
		if len(workflowInstance.Jobs) == 0 {
			return true, nil
		}

		for jobName, job := range workflowInstance.Jobs {
			if len(job.Env) == 0 {
				continue
			}

			for envName := range job.Env {
				m := casematch.Match(envName, confValue)
				if !m {
					chErrors <- glitch.Glitch{
						Path:     workflowInstance.Path,
						Name:     workflowInstance.DisplayName,
						Type:     rule.DotGithubFileTypeWorkflow,
						ErrText:  fmt.Sprintf("job '%s' env '%s' must be %s", jobName, envName, confValue),
						RuleName: r.ConfigName(0),
					}
				}
			}
		}
	case WorkflowFieldJobStepEnv:
		for jobName, job := range workflowInstance.Jobs {
			for stepIdx, step := range job.Steps {
				if len(step.Env) == 0 {
					continue
				}

				for envName := range step.Env {
					m := casematch.Match(envName, confValue)
					if !m {
						chErrors <- glitch.Glitch{
							Path:     workflowInstance.Path,
							Name:     workflowInstance.DisplayName,
							Type:     rule.DotGithubFileTypeWorkflow,
							ErrText:  fmt.Sprintf("job '%s' step %d env '%s' must be %s", jobName, stepIdx, envName, confValue),
							RuleName: r.ConfigName(0),
						}

						compliant = false
					}
				}
			}
		}
	case WorkflowFieldReferencedVariable:
		varTypes := []string{"env", "vars", "secrets"}
		for _, v := range varTypes {
			re := regexp.MustCompile(fmt.Sprintf("\\${{[ ]*%s\\.([a-zA-Z0-9\\-_]+)[ ]*}}", v))

			found := re.FindAllSubmatch(workflowInstance.Raw, -1)
			for _, refVar := range found {
				m := casematch.Match(string(refVar[1]), confValue)
				if !m {
					chErrors <- glitch.Glitch{
						Path:     workflowInstance.Path,
						Name:     workflowInstance.DisplayName,
						Type:     rule.DotGithubFileTypeWorkflow,
						ErrText:  fmt.Sprintf("calls a variable '%s' that must be %s", string(refVar[1]), confValue),
						RuleName: r.ConfigName(0),
					}

					compliant = false
				}
			}
		}
	case WorkflowFieldDispatchInputName:
		if workflowInstance.On == nil ||
			workflowInstance.On.WorkflowDispatch == nil ||
			len(workflowInstance.On.WorkflowDispatch.Inputs) == 0 {
			return true, nil
		}

		for inputName := range workflowInstance.On.WorkflowDispatch.Inputs {
			m := casematch.Match(inputName, confValue)
			if !m {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("call input '%s' name must be %s", inputName, confValue),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	case WorkflowFieldCallInputName:
		if workflowInstance.On == nil ||
			workflowInstance.On.WorkflowCall == nil ||
			len(workflowInstance.On.WorkflowCall.Inputs) == 0 {
			return true, nil
		}

		for inputName := range workflowInstance.On.WorkflowCall.Inputs {
			m := casematch.Match(inputName, confValue)
			if !m {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("dispatch input '%s' name must be %s", inputName, confValue),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	case WorkflowFieldJobName:
		if len(workflowInstance.Jobs) == 0 {
			return true, nil
		}

		for jobName := range workflowInstance.Jobs {
			m := casematch.Match(jobName, confValue)
			if !m {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("job '%s' name must be %s", jobName, confValue),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	}

	return compliant, nil
}

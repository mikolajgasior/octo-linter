package dependencies

import (
	"fmt"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// WorkflowReferencedVariableExistsInFile checks if called variables and secrets exist.
// This rule requires a list of variables and secrets to be checked against.
type WorkflowReferencedVariableExistsInFile struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r WorkflowReferencedVariableExistsInFile) ConfigName(int) string {
	return "dependencies__workflow_referenced_variable_must_exists_in_attached_file"
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r WorkflowReferencedVariableExistsInFile) FileType() int {
	return rule.DotGithubFileTypeWorkflow
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r WorkflowReferencedVariableExistsInFile) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
func (r WorkflowReferencedVariableExistsInFile) Lint(
	conf interface{},
	file dotgithub.File,
	dotGithub *dotgithub.DotGithub,
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

	compliant := true

	varTypes := []string{"vars", "secrets"}
	for _, varType := range varTypes {
		re := regexpRefs[varType]
		if re == nil {
			continue
		}

		found := re.FindAllSubmatch(workflowInstance.Raw, -1)
		for _, refVar := range found {
			if varType == "vars" && len(dotGithub.Vars) > 0 &&
				!dotGithub.IsVarExist(string(refVar[1])) {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("calls a variable '%s' that does not exist in the vars file", string(refVar[1])),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}

			if varType == "secrets" && len(dotGithub.Secrets) > 0 &&
				!dotGithub.IsSecretExist(string(refVar[1])) {
				chErrors <- glitch.Glitch{
					Path:     workflowInstance.Path,
					Name:     workflowInstance.DisplayName,
					Type:     rule.DotGithubFileTypeWorkflow,
					ErrText:  fmt.Sprintf("calls a secret '%s' that does not exist in the secrets file", string(refVar[1])),
					RuleName: r.ConfigName(0),
				}

				compliant = false
			}
		}
	}

	return compliant, nil
}

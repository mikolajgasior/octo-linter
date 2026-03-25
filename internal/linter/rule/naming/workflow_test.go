package naming

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestWorkflowValidate(t *testing.T) {
	t.Parallel()

	rule := Workflow{}

	confBad := "some string"

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("Workflow.Validate should return error when conf is %v", confBad)
	}

	confGood := ValueCamelCase

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf(
			"Workflow.Validate should not return error (%s) when conf is %v",
			err.Error(),
			confGood,
		)
	}
}

func TestWorkflowNotCompliant(t *testing.T) {
	t.Parallel()

	for field, conf := range map[int]string{
		WorkflowFieldEnv:                ValueAllCaps,
		WorkflowFieldJobEnv:             ValueAllCaps,
		WorkflowFieldJobStepEnv:         ValueAllCaps,
		WorkflowFieldReferencedVariable: ValueAllCaps,
		WorkflowFieldDispatchInputName:  ValueDashCase,
		WorkflowFieldCallInputName:      ValueDashCase,
		WorkflowFieldJobName:            ValueDashCase,
	} {
		rule := Workflow{
			Field: field,
		}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if compliant {
				t.Errorf(
					"Workflow.Lint should return false when workflow field %d does not follow naming convention of '%s'",
					field,
					conf,
				)
			}

			if err != nil {
				t.Errorf("Workflow.Lint failed with an error: %s", err.Error())
			}

			if len(ruleErrors) != 2 {
				t.Errorf(
					"Workflow.Lint should send 2 errors over the channel, got [%s]",
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		ruletest.Workflow(d, "naming-workflow", fn)
	}
}

func TestWorkflowCompliant(t *testing.T) {
	t.Parallel()

	for field, conf := range map[int]string{
		WorkflowFieldEnv:                ValueAllCaps,
		WorkflowFieldJobEnv:             ValueAllCaps,
		WorkflowFieldJobStepEnv:         ValueAllCaps,
		WorkflowFieldReferencedVariable: ValueAllCaps,
		WorkflowFieldDispatchInputName:  ValueDashCase,
		WorkflowFieldCallInputName:      ValueDashCase,
		WorkflowFieldJobName:            ValueDashCase,
	} {
		rule := Workflow{
			Field: field,
		}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if !compliant {
				t.Errorf(
					"Workflow.Lint should return true when workflow field %d follows naming convention of '%s'",
					field,
					conf,
				)
			}

			if err != nil {
				t.Errorf("Workflow.Lint failed with an error: %s", err.Error())
			}

			if len(ruleErrors) != 0 {
				t.Errorf(
					"Workflow.Lint should not send any errors over the channel, got [%s]",
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		ruletest.Workflow(d, "valid-workflow.yml", fn)
	}
}

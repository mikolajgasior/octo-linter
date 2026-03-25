package required

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestWorkflowValidate(t *testing.T) {
	t.Parallel()

	rule := Workflow{
		Field: WorkflowFieldWorkflow,
	}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("Workflow.Validate should return error when conf is not []string")
	}

	confGood := []interface{}{ValueName}

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("Workflow.Validate should not return error when conf is []string")
	}

	for _, f := range []int{WorkflowFieldDispatchInput, WorkflowFieldCallInput} {
		rule = Workflow{
			Field: f,
		}

		confBad2 := []interface{}{ValueName, ValueDesc}

		err = rule.Validate(confBad2)
		if err == nil {
			t.Errorf("Workflow.Validate should return error when conf contains invalid values")
		}

		confGood2 := []interface{}{ValueDesc}

		err = rule.Validate(confGood2)
		if err != nil {
			t.Errorf("Workflow.Validate should not return error when conf contains valid values")
		}
	}
}

func TestWorkflowFieldWorkflowNotCompliant(t *testing.T) {
	t.Parallel()

	rule := Workflow{
		Field: WorkflowFieldWorkflow,
	}
	conf := []interface{}{"name"}
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf("Workflow.Lint should return false when workflow does not have a 'name' field")
		}

		if err != nil {
			t.Errorf("Workflow.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) == 0 {
			t.Errorf("Workflow.Lint should send an error over the channel")
		}
	}

	ruletest.Workflow(d, "required-workflow.yml", fn)
}

func TestWorkflowFieldCallInputDispatchInputNotCompliant(t *testing.T) {
	t.Parallel()

	for _, field := range []int{WorkflowFieldDispatchInput, WorkflowFieldCallInput} {
		rule := Workflow{
			Field: field,
		}
		conf := []interface{}{ValueDesc}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if compliant {
				t.Errorf(
					"Workflow.Lint should return false when workflow field %d does not have a 'description' field",
					field,
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

		ruletest.Workflow(d, "required-workflow.yml", fn)
	}
}

func TestWorkflowFieldWorkflowCompliant(t *testing.T) {
	t.Parallel()

	rule := Workflow{
		Field: WorkflowFieldWorkflow,
	}
	conf := []interface{}{ValueName}
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf("Workflow.Lint should return true when workflow has a 'name' field")
		}

		if err != nil {
			t.Errorf("Workflow.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 0 {
			t.Errorf(
				"Workflow.Lint should not send any error over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Workflow(d, "valid-workflow.yml", fn)
}

func TestWorkflowFieldCallInputDispatchInputCompliant(t *testing.T) {
	t.Parallel()

	for _, field := range []int{WorkflowFieldDispatchInput, WorkflowFieldCallInput} {
		rule := Workflow{
			Field: field,
		}
		conf := []interface{}{ValueDesc}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if !compliant {
				t.Errorf(
					"Workflow.Lint should return true when workflow field %d has a 'description' field",
					field,
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

package dependencies

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestReferencedInputExistsValidate(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("ReferencedInputExists.Validate should return error when conf is not bool")
	}

	confGood := true

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("ReferencedInputExists.Validate should not return error when conf is bool")
	}
}

func TestReferencedInputExistsActionNotCompliant(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{
		FileTypeRequired: "action",
	}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"ReferencedInputExists.Lint should return false when there are referenced to non-existing inputs and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("ReferencedInputExists.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 2 {
			t.Errorf(
				"ReferencedInputExists.Lint should send 2 errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "dependencies-referenced-input-exists", fn)
}

func TestReferencedInputExistsActionCompliant(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{
		FileTypeRequired: "action",
	}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"ReferencedInputExists.Lint should return true when there are no referenced to non-existing inputs and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("ReferencedInputExists.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 0 {
			t.Errorf(
				"ReferencedInputExists.Lint should not send any errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "valid-action", fn)
}

func TestReferencedInputExistsWorkflowNotCompliant(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{
		FileTypeRequired: "workflow",
	}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"ReferencedInputExists.Lint should return false when there are referenced to non-existing inputs and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("ReferencedInputExists.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 2 {
			t.Errorf(
				"ReferencedInputExists.Lint should send 2 errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Workflow(d, "dependencies-referenced-input-exists.yml", fn)
}

func TestReferencedInputExistsWorkflowCompliant(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{
		FileTypeRequired: "workflow",
	}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"ReferencedInputExists.Lint should return true when there are no referenced to non-existing inputs and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("ReferencedInputExists.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 0 {
			t.Errorf(
				"ReferencedInputExists.Lint should not send any errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Workflow(d, "valid-workflow.yml", fn)
}

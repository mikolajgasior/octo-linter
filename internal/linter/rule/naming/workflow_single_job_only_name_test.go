package naming

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

const (
	// TestOnlyJobName defines the name of the only existing job's name.
	TestOnlyJobName = "main"
)

func TestWorkflowSingleJobOnlyNameValidate(t *testing.T) {
	t.Parallel()

	rule := WorkflowSingleJobOnlyName{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("Workflow.Validate should return error when conf is not string")
	}

	confGood := TestOnlyJobName

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("Workflow.Validate should not return error when conf is string")
	}
}

func TestWorkflowSingleJobOnlyNameNotCompliant(t *testing.T) {
	t.Parallel()

	rule := WorkflowSingleJobOnlyName{}
	conf := TestOnlyJobName
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"Workflow.Lint should return false when workflow has only job and its name is not '%s'",
				conf,
			)
		}

		if err != nil {
			t.Errorf("Workflow.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) == 0 {
			t.Errorf("Workflow.Lint should send an error over the channel")
		}
	}

	ruletest.Workflow(d, "naming-workflow-single-job-only-name.yml", fn)
}

func TestWorkflowSingleJobOnlyNameCompliant(t *testing.T) {
	t.Parallel()

	rule := WorkflowSingleJobOnlyName{}
	conf := TestOnlyJobName
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"Workflow.Lint should return true when workflow has only job and its name is '%s'",
				conf,
			)
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

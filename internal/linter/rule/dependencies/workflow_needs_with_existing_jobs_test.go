package dependencies

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestWorkflowNeedsWithExistingJobsValidate(t *testing.T) {
	t.Parallel()

	rule := WorkflowNeedsWithExistingJobs{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("WorkflowNeedsWithExistingJobs.Validate should return error when conf is not bool")
	}

	confGood := true

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("WorkflowNeedsWithExistingJobs.Validate should not return error when conf is bool")
	}
}

func TestWorkflowNeedsWithExistingJobsNotCompliant(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"WorkflowNeedsWithExistingJobs.Lint should return false when invalid dependencies between jobs and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("WorkflowNeedsWithExistingJobs.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 2 {
			t.Errorf(
				"WorkflowNeedsWithExistingJobs.Lint should send 2 errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "dependencies-workflow-needs-with-existing-jobs", fn)
}

func TestWorkflowNeedsWithExistingJobsCompliant(t *testing.T) {
	t.Parallel()

	rule := ReferencedInputExists{}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"WorkflowNeedsWithExistingJobs.Lint should return true dependencies between jobs are valid and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("WorkflowNeedsWithExistingJobs.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 0 {
			t.Errorf(
				"WorkflowNeedsWithExistingJobs.Lint should not send any errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Workflow(d, "valid-workflow.yml", fn)
}

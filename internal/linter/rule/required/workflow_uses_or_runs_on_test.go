package required

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestWorkflowUsesOrRunsOnValidate(t *testing.T) {
	t.Parallel()

	rule := WorkflowUsesOrRunsOn{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("WorkflowUsesOrRunsOn.Validate should return error when conf is not bool")
	}

	confGood := true

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("WorkflowUsesOrRunsOn.Validate should not return error when conf is bool")
	}
}

func TestWorkflowUsesOrRunsOnNotCompliant(t *testing.T) {
	t.Parallel()

	rule := WorkflowUsesOrRunsOn{}
	conf := true
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"WorkflowUsesOrRunsOn.Lint should return false when a job does not have 'uses' or 'runs-on'",
			)
		}

		if err != nil {
			t.Errorf("WorkflowUsesOrRunsOn.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) == 0 {
			t.Errorf("WorkflowUsesOrRunsOn.Lint should send an error over the channel")
		}
	}

	ruletest.Workflow(d, "required-workflow-uses-or-runs-on.yml", fn)
}

func TestWorkflowUsesOrRunsOnCompliant(t *testing.T) {
	t.Parallel()

	rule := WorkflowUsesOrRunsOn{}
	conf := true
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"WorkflowUsesOrRunsOn.Lint should return true when all the jobs have either 'uses' or 'runs-on'",
			)
		}

		if err != nil {
			t.Errorf("WorkflowUsesOrRunsOn.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) > 0 {
			t.Errorf(
				"WorkflowUsesOrRunsOn.Lint should not send any error over the channel, sent %s",
				strings.Join(ruleErrors, "|"),
			)
		}
	}

	ruletest.Workflow(d, "valid-workflow.yml", fn)
}

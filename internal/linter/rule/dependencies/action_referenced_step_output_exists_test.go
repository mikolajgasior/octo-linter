package dependencies

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestActionReferencedStepOutputExistsValidate(t *testing.T) {
	t.Parallel()

	rule := ActionReferencedStepOutputExists{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf(
			"ActionReferencedStepOutputExists.Validate should return error when conf is not bool",
		)
	}

	confGood := true

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf(
			"ActionReferencedStepOutputExists.Validate should not return error when conf is bool",
		)
	}
}

func TestActionReferencedStepOutputExistsNotCompliant(t *testing.T) {
	t.Parallel()

	rule := ActionReferencedStepOutputExists{}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"ActionReferencedStepOutputExists.Lint should return false when there are invalid step outputs used in it"+
					" and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("ActionReferencedStepOutputExists.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 4 {
			t.Errorf(
				"ActionReferencedStepOutputExists.Lint should send 4 errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "dependencies-action-referenced-step-output-exists", fn)
}

func TestActionCompliant(t *testing.T) {
	t.Parallel()

	rule := ActionReferencedStepOutputExists{}
	d := ruletest.GetDotGithub()
	conf := true

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"ActionReferencedStepOutputExists.Lint should return true when action does not call invalid step outputs in"+
					" it and conf is %v",
				conf,
			)
		}

		if err != nil {
			t.Errorf("ActionReferencedStepOutputExists.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 0 {
			t.Errorf(
				"ActionReferencedStepOutputExists.Lint should not send any errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "valid-action", fn)
}

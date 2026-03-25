package refvars

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestNotOneWordValidate(t *testing.T) {
	t.Parallel()

	rule := NotOneWord{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("NotOneWord.Validate should return error when conf is not bool")
	}

	confGood := true

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("NotOneWord.Validate should not return error when conf is bool")
	}
}

func TestNotOneWordNotCompliant(t *testing.T) {
	t.Parallel()

	rule := NotOneWord{}
	conf := true
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"NotOneWord.Lint should return false when there is a reference to a 'one-word' variable",
			)
		}

		if err != nil {
			t.Errorf("NotOneWord.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) == 0 {
			t.Errorf("NotOneWord.Lint should send an error over the channel")
		}
	}

	ruletest.Action(d, "refvars-not-one-word", fn)
	ruletest.Workflow(d, "refvars-not-one-word.yml", fn)
}

func TestNotOneWordCompliant(t *testing.T) {
	t.Parallel()

	rule := NotOneWord{}
	conf := true
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"NotOneWord.Lint should return true when there are not any vars that are one word",
			)
		}

		if err != nil {
			t.Errorf("NotOneWord.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) > 0 {
			t.Errorf(
				"NotOneWord.Lint should not send any error over the channel, sent %s",
				strings.Join(ruleErrors, "|"),
			)
		}
	}

	ruletest.Action(d, "valid-action", fn)
	ruletest.Workflow(d, "valid-workflow.yml", fn)
}

package refvars

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestNotInDoubleQuotesValidate(t *testing.T) {
	t.Parallel()

	rule := NotInDoubleQuotes{}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("NotInDoubleQuotes.Validate should return error when conf is not bool")
	}

	confGood := true

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("NotInDoubleQuotes.Validate should not return error when conf is bool")
	}
}

func TestNotInDoubleQuotesNotCompliant(t *testing.T) {
	t.Parallel()

	for _, fileTypeRequired := range []string{"action", "workflow"} {
		rule := NotInDoubleQuotes{
			FileTypeRequired: fileTypeRequired,
		}
		conf := true
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if compliant {
				t.Errorf(
					"NotInDoubleQuotes.Lint should return false when there is a variable in double quotes",
				)
			}

			if err != nil {
				t.Errorf("NotInDoubleQuotes.Lint failed with an error: %s", err.Error())
			}

			if len(ruleErrors) == 0 {
				t.Errorf("NotInDoubleQuotes.Lint should send an error over the channel")
			}
		}

		if fileTypeRequired == "action" {
			ruletest.Action(d, "refvars-not-in-double-quotes", fn)
		} else {
			ruletest.Workflow(d, "refvars-not-in-double-quotes.yml", fn)
		}
	}
}

func TestNotInDoubleQuotesCompliant(t *testing.T) {
	t.Parallel()

	for _, fileTypeRequired := range []string{"action", "workflow"} {
		rule := NotInDoubleQuotes{
			FileTypeRequired: fileTypeRequired,
		}
		conf := true
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if !compliant {
				t.Errorf(
					"NotInDoubleQuotes.Lint should return true when there are not any vars that are in double quotes",
				)
			}

			if err != nil {
				t.Errorf("NotInDoubleQuotes.Lint failed with an error: %s", err.Error())
			}

			if len(ruleErrors) > 0 {
				t.Errorf(
					"NotInDoubleQuotes.Lint should not send any error over the channel, sent %s",
					strings.Join(ruleErrors, "|"),
				)
			}
		}

		if fileTypeRequired == "action" {
			ruletest.Action(d, "valid-action", fn)
		} else {
			ruletest.Workflow(d, "valid-workflow.yml", fn)
		}
	}
}

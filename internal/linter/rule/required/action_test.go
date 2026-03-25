package required

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestActionValidate(t *testing.T) {
	t.Parallel()

	rule := Action{
		Field: ActionFieldAction,
	}

	confBad := 4

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("Action.Validate should return error when conf is not []string")
	}

	confGood := []interface{}{ValueName, ValueDesc}

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf("Action.Validate should not return error when conf is []string")
	}

	for _, f := range []int{ActionFieldInput, ActionFieldOutput} {
		rule = Action{
			Field: f,
		}

		confBad2 := []interface{}{ValueName, ValueDesc}

		err = rule.Validate(confBad2)
		if err == nil {
			t.Errorf("Action.Validate should return error when conf contains invalid values")
		}

		confGood2 := []interface{}{ValueDesc}

		err = rule.Validate(confGood2)
		if err != nil {
			t.Errorf("Action.Validate should not return error when conf contains valid values")
		}
	}
}

func TestActionFieldActionNotCompliant(t *testing.T) {
	t.Parallel()

	rule := Action{
		Field: ActionFieldAction,
	}
	conf := []interface{}{ValueName, ValueDesc}
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if compliant {
			t.Errorf(
				"Action.Lint should return false when action does not have a 'name' and/or 'description' field",
			)
		}

		if err != nil {
			t.Errorf("Action.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 2 {
			t.Errorf(
				"Action.Lint should send 2 errors over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "required-action", fn)
}

func TestActionFieldInputOutputNotCompliant(t *testing.T) {
	t.Parallel()

	for _, field := range []int{ActionFieldInput, ActionFieldOutput} {
		rule := Action{
			Field: field,
		}
		conf := []interface{}{ValueDesc}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if compliant {
				t.Errorf(
					"Action.Lint should return false when action field %d does not have a 'description' field",
					field,
				)
			}

			if err != nil {
				t.Errorf("Action.Lint failed with an error: %s", err.Error())
			}

			if len(ruleErrors) != 2 {
				t.Errorf(
					"Action.Lint should send 2 errors over the channel, got [%s]",
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		ruletest.Action(d, "required-action", fn)
	}
}

func TestActionFieldActionCompliant(t *testing.T) {
	t.Parallel()

	rule := Action{
		Field: ActionFieldAction,
	}
	conf := []interface{}{ValueName, ValueDesc}
	d := ruletest.GetDotGithub()

	fn := func(f dotgithub.File, _ string) {
		compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
		if !compliant {
			t.Errorf(
				"Action.Lint should return true when action has both 'name' and 'description' field",
			)
		}

		if err != nil {
			t.Errorf("Action.Lint failed with an error: %s", err.Error())
		}

		if len(ruleErrors) != 0 {
			t.Errorf(
				"Action.Lint should not send any error over the channel, got [%s]",
				strings.Join(ruleErrors, "\n"),
			)
		}
	}

	ruletest.Action(d, "valid-action", fn)
}

func TestActionFieldInputOutputCompliant(t *testing.T) {
	t.Parallel()

	for _, field := range []int{ActionFieldInput, ActionFieldOutput} {
		rule := Action{
			Field: field,
		}
		conf := []interface{}{ValueDesc}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, _ string) {
			compliant, ruleErrors, err := ruletest.Lint(2, rule, conf, f, d)
			if !compliant {
				t.Errorf(
					"Action.Lint should return true when action field %d has a 'description' field",
					field,
				)
			}

			if err != nil {
				t.Errorf("Action.Lint failed with an error: %s", err.Error())
			}

			if len(ruleErrors) != 0 {
				t.Errorf(
					"Action.Lint should not send any errors over the channel, got [%s]",
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		ruletest.Action(d, "valid-action", fn)
	}
}

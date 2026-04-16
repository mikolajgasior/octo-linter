package usedactions

import (
	"strings"
	"testing"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/ruletest"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

func TestExistsValidate(t *testing.T) {
	t.Parallel()

	rule := Exists{}

	confBad := []interface{}{"something", "something2"}

	err := rule.Validate(confBad)
	if err == nil {
		t.Errorf("Exists.Validate should return error when conf is %v", confBad)
	}

	confGood := []interface{}{"local", "external"}

	err = rule.Validate(confGood)
	if err != nil {
		t.Errorf(
			"Exists.Validate should not return error (%s) when conf is %v",
			err.Error(),
			confGood,
		)
	}
}

func TestLocal(t *testing.T) {
	t.Parallel()

	for _, fileTypeRequired := range []string{"action", "workflow"} {
		rule := Exists{
			FileTypeRequired: fileTypeRequired,
		}
		conf := []interface{}{"local"}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, n string) {
			compliant, ruleErrors, err := ruletest.Lint(3, rule, conf, f, d)
			if compliant {
				t.Errorf("Exists.Lint on %s should return false when conf is %v", n, conf)
			}

			if err != nil {
				t.Errorf("Exists.Lint on %s failed with an error: %s", n, err.Error())
			}

			if len(ruleErrors) != 2 {
				t.Errorf(
					"Exists.Lint on %s should send 2 errors over the channel not [%s]",
					n,
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		if fileTypeRequired == "action" {
			ruletest.Action(d, "usedactions-exists", fn)
		} else {
			ruletest.Workflow(d, "usedactions-exists.yml", fn)
		}
	}
}

func TestExternal(t *testing.T) {
	t.Parallel()

	for _, fileTypeRequired := range []string{"action", "workflow"} {
		rule := Exists{
			FileTypeRequired: fileTypeRequired,
		}
		conf := []interface{}{"external"}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, n string) {
			compliant, ruleErrors, err := ruletest.Lint(3, rule, conf, f, d)
			if compliant {
				t.Errorf("Exists.Lint on %s should return false when conf is %v", n, conf)
			}

			if err != nil {
				t.Errorf("Exists.Lint on %s failed with an error: %s", n, err.Error())
			}

			if len(ruleErrors) != 2 {
				t.Errorf(
					"Exists.Lint on %s should send 2 errors over the channel not [%s]",
					n,
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		if fileTypeRequired == "action" {
			ruletest.Action(d, "usedactions-exists", fn)
		} else {
			ruletest.Workflow(d, "usedactions-exists.yml", fn)
		}
	}
}

func TestLocalExternal(t *testing.T) {
	t.Parallel()

	for _, fileTypeRequired := range []string{"action", "workflow"} {
		rule := Exists{
			FileTypeRequired: fileTypeRequired,
		}
		conf := []interface{}{"local", "external"}
		d := ruletest.GetDotGithub()

		fn := func(f dotgithub.File, n string) {
			compliant, ruleErrors, err := ruletest.Lint(3, rule, conf, f, d)
			if compliant {
				t.Errorf("Exists.Lint on %s should return false when conf is %v", n, conf)
			}

			if err != nil {
				t.Errorf("Exists.Lint on %s failed with an error: %s", n, err.Error())
			}

			if len(ruleErrors) != 4 {
				t.Errorf(
					"Exists.Lint on %s should send 4 errors over the channel not [%s]",
					n,
					strings.Join(ruleErrors, "\n"),
				)
			}
		}

		if fileTypeRequired == "action" {
			ruletest.Action(d, "usedactions-exists", fn)
		} else {
			ruletest.Workflow(d, "usedactions-exists.yml", fn)
		}
	}
}

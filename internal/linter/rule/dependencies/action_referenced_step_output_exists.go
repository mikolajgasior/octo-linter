package dependencies

import (
	"fmt"
	"strings"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

// ActionReferencedStepOutputExists checks whether references to step outputs correspond to outputs defined in
// preceding steps. During execution, referencing a non-existent step output results in an empty string.
type ActionReferencedStepOutputExists struct{}

// ConfigName returns the name of the rule as defined in the configuration file.
func (r ActionReferencedStepOutputExists) ConfigName(int) string {
	return "dependencies__action_referenced_step_output_must_exist"
}

// FileType returns an integer that specifies the file types (action and/or workflow) the rule targets.
func (r ActionReferencedStepOutputExists) FileType() int {
	return rule.DotGithubFileTypeAction
}

// Validate checks whether the given value is valid for this rule's configuration.
func (r ActionReferencedStepOutputExists) Validate(conf interface{}) error {
	_, ok := conf.(bool)
	if !ok {
		return errValueNotBool
	}

	return nil
}

// Lint runs a rule with the specified configuration on a dotgithub.File (action or workflow),
// reports any errors via the given channel, and returns whether the file is compliant.
//
//nolint:gocognit,funlen
func (r ActionReferencedStepOutputExists) Lint(
	conf interface{},
	file dotgithub.File,
	dotGithub *dotgithub.DotGithub,
	chErrors chan<- glitch.Glitch,
) (bool, error) {
	confValue, confIsBool := conf.(bool)
	if !confIsBool {
		return false, errValueNotBool
	}

	if file.GetType() != rule.DotGithubFileTypeAction || !confValue {
		return true, nil
	}

	actionInstance, ok := file.(*action.Action)
	if !ok {
		return false, errFileInvalidType
	}

	compliant := true

	found := regexpStepOutput.FindAllSubmatch(actionInstance.Raw, -1)
	for _, foundStepOutput := range found {
		stepName := string(foundStepOutput[1])
		outputName := string(foundStepOutput[2])

		if actionInstance.Runs == nil {
			chErrors <- glitch.Glitch{
				Path:     actionInstance.Path,
				Name:     actionInstance.DirName,
				Type:     rule.DotGithubFileTypeAction,
				ErrText:  fmt.Sprintf("calls a step output '%s' but 'runs' does not exist", stepName),
				RuleName: r.ConfigName(0),
			}

			compliant = false

			continue
		}

		step := actionInstance.Runs.GetStep(string(foundStepOutput[1]))
		if step == nil {
			chErrors <- glitch.Glitch{
				Path:     actionInstance.Path,
				Name:     actionInstance.DirName,
				Type:     rule.DotGithubFileTypeAction,
				ErrText:  fmt.Sprintf("calls a step '%s' output '%s' but step does not exist", stepName, outputName),
				RuleName: r.ConfigName(0),
			}

			compliant = false

			continue
		}

		foundOutput := false

		// search in 'run' when there is no 'uses'
		if step.Uses == "" && step.Run != "" {
			foundEchoLines := regexpAppendToGithubOutput.FindAllSubmatch([]byte(step.Run), -1)
			for _, foundEchoLine := range foundEchoLines {
				if outputName == string(foundEchoLine[1]) {
					foundOutput = true
				}
			}

			if !foundOutput {
				chErrors <- glitch.Glitch{
					Path:     actionInstance.Path,
					Name:     actionInstance.DirName,
					Type:     rule.DotGithubFileTypeAction,
					ErrText:  fmt.Sprintf("calls a step '%s' output '%s' that does not exist", stepName, outputName),
					RuleName: r.ConfigName(0),
				}

				compliant = false

				continue
			}
		}

		if foundOutput {
			continue
		}

		var foundAction *action.Action
		// local action
		if regexpLocal.MatchString(step.Uses) {
			actionName := strings.ReplaceAll(step.Uses, "./.github/actions/", "")
			foundAction = dotGithub.GetAction(actionName)
		}
		// external action
		if regexpExternal.MatchString(step.Uses) {
			foundAction = dotGithub.GetExternalAction(step.Uses)
		}

		if foundAction == nil {
			chErrors <- glitch.Glitch{
				Path:     actionInstance.Path,
				Name:     actionInstance.DirName,
				Type:     rule.DotGithubFileTypeAction,
				ErrText:  fmt.Sprintf("calls a step '%s' output '%s' on action that does not exist", stepName, outputName),
				RuleName: r.ConfigName(0),
			}

			compliant = false

			continue
		}

		for duaOutputName := range foundAction.Outputs {
			if duaOutputName == outputName {
				foundOutput = true
			}
		}

		if !foundOutput {
			chErrors <- glitch.Glitch{
				Path:     actionInstance.Path,
				Name:     actionInstance.DirName,
				Type:     rule.DotGithubFileTypeAction,
				ErrText:  fmt.Sprintf("calls step '%s' output '%s' on action and that output does not exist", stepName, outputName),
				RuleName: r.ConfigName(0),
			}

			compliant = false

			continue
		}
	}

	return compliant, nil
}

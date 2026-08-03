package linter

import (
	"octo-linter/internal/dotgithub"
	"octo-linter/internal/linter/glitch"
	"octo-linter/internal/linter/rule"
)

// Job represents a single run of a rule against a .github file (action or workflow).
type Job struct {
	rule      rule.Rule
	file      dotgithub.File
	dotGithub *dotgithub.DotGithub
	isError   bool
	value     interface{}
}

// Run executes the Job, routing any lint findings to the appropriate channel.
func (j *Job) Run(chWarnings chan<- glitch.Glitch, chErrors chan<- glitch.Glitch) (bool, error) {
	if j.isError {
		return j.rule.Lint(j.value, j.file, j.dotGithub, chErrors)
	}

	return j.rule.Lint(j.value, j.file, j.dotGithub, chWarnings)
}

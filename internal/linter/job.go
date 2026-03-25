package linter

import (
	"errors"
	"fmt"
	"time"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

const (
	// SecondsJobTimeout sets the job timeout in seconds.
	SecondsJobTimeout = 10
)

var (
	errLintTimeout = errors.New("lint timeout")
	errLintError   = errors.New("lint error")
)

func errRuleLintTimeout(name string) error {
	return fmt.Errorf("%w: %s", errLintTimeout, name)
}

func errRuleLintError(err error) error {
	return fmt.Errorf("%w: %s", errLintError, err.Error())
}

// Job represents a single run of a rule against a .github file (action or workflow).
type Job struct {
	rule      rule.Rule
	file      dotgithub.File
	dotGithub *dotgithub.DotGithub
	isError   bool
	value     interface{}
}

// Run execute the Job and sends any errors or warnings to specified channels.
func (j *Job) Run(chWarnings chan<- glitch.Glitch, chErrors chan<- glitch.Glitch) (bool, error) {
	compliant := true

	var err error

	done := make(chan struct{})
	timer := time.NewTimer(SecondsJobTimeout * time.Second)

	go func() {
		if j.isError {
			compliant, err = j.rule.Lint(j.value, j.file, j.dotGithub, chErrors)
		} else {
			compliant, err = j.rule.Lint(j.value, j.file, j.dotGithub, chWarnings)
		}

		close(done)
	}()

	select {
	case <-timer.C:
		return false, errRuleLintTimeout(j.rule.ConfigName(j.file.GetType()))
	case <-done:
		if err != nil {
			return compliant, errRuleLintError(err)
		}

		return compliant, nil
	}
}

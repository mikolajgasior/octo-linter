// Package rule contains a definition of linting rule.
package rule

import (
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

const (
	// DotGithubFileTypeAction represents the action file type. Used in a bitmask and must be a power of 2.
	DotGithubFileTypeAction = 1
	// DotGithubFileTypeWorkflow represents the workflow file type. Used in a bitmask and must be a power of 2.
	DotGithubFileTypeWorkflow = 2
)

// Rule represents a rule.
type Rule interface {
	Validate(conf interface{}) error
	Lint(
		config interface{},
		f dotgithub.File,
		d *dotgithub.DotGithub,
		chErrors chan<- glitch.Glitch,
	) (bool, error)
	ConfigName(fileType int) string
	FileType() int
}

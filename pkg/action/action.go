// Package action contains code related to GitHub Actions action.
package action

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"

	"gopkg.in/yaml.v2"
)

const (
	// DotGithubFileTypeAction represents the action file type. Used in a bitmask and must be a power of 2.
	DotGithubFileTypeAction = 1
)

// Action represents a GitHub Actions' action parsed from a YAML file.
type Action struct {
	Path           string
	Raw            []byte
	DirName        string
	Name           string             `yaml:"name"`
	Description    string             `yaml:"description"`
	Inputs         map[string]*Input  `yaml:"inputs"`
	Outputs        map[string]*Output `yaml:"outputs"`
	DynamicOutputs []*regexp.Regexp   `yaml:"-"`
	Runs           *Runs              `yaml:"runs"`
}

// Unmarshal parses YAML from a file in struct's Path or from struct's Raw field.
func (a *Action) Unmarshal(fromRaw bool) error {
	if !fromRaw {
		slog.Debug(
			"reading action file",
			slog.String("path", a.Path),
		)

		b, err := os.ReadFile(a.Path)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %w", a.Path, err)
		}

		a.Raw = b
	}

	err := yaml.Unmarshal(a.Raw, &a)
	if err != nil {
		return fmt.Errorf("cannot unmarshal file %s: %w", a.Path, err)
	}

	if a.Runs != nil {
		a.Runs.SetParentType("action")
	}

	return nil
}

// GetType returns the int value representing the action file type. See dotgithub.File interface.
func (a *Action) GetType() int {
	return DotGithubFileTypeAction
}

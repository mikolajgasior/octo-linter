package linter

import "regexp"

type Overrides struct {
	ExternalActionsOutputsConfig map[string][]string         `yaml:"external_actions_outputs,omitempty"`
	ExternalActionsOutputs       map[string][]*regexp.Regexp `yaml:"-"`
	ExternalActionsPaths         map[string]string           `yaml:"external_actions_paths,omitempty"`
}

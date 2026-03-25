// Package linter contains code related to octo-linter configuration.
package linter

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"gopkg.in/yaml.v2"
)

//go:embed dotgithub.yml
var defaultConfig []byte

// Config represents the configuration file.
type Config struct {
	Version     string                            `yaml:"version"`
	RulesConfig map[string]map[string]interface{} `yaml:"rules"`
	Rules       []rule.Rule                       `yaml:"-"`
	Values      []interface{}                     `yaml:"-"`
	WarningOnly map[string]struct{}               `yaml:"-"`
}

// GetDefaultConfig returns a default configuration file.
func GetDefaultConfig() []byte {
	return defaultConfig
}

// ReadFile parses configuration from a specified file.
func (cfg *Config) ReadFile(path string) error {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	err = cfg.readBytesAndValidate(b)
	if err != nil {
		return fmt.Errorf("error reading and/or validating config file %s: %w", path, err)
	}

	return nil
}

// ReadDefaultFile sets the Config from a default configuration file.
func (cfg *Config) ReadDefaultFile() error {
	err := cfg.readBytesAndValidate(defaultConfig)
	if err != nil {
		return fmt.Errorf("error reading and/or validating default config file: %w", err)
	}

	return nil
}

// IsError checks if rule has been set to have a status of error.
func (cfg *Config) IsError(rule string) bool {
	_, isWarn := cfg.WarningOnly[rule]

	return !isWarn
}

func (cfg *Config) readBytesAndValidate(b []byte) error {
	cfg.Rules = make([]rule.Rule, 0)
	cfg.Values = make([]interface{}, 0)

	err := yaml.Unmarshal(b, &cfg)
	if err != nil {
		return fmt.Errorf("error unmarshalling: %w", err)
	}

	cfg.WarningOnly = make(map[string]struct{})

	for ruleGroupName, ruleGroup := range cfg.RulesConfig {
		warningOnly := make(map[string]struct{})

		// Parse out rules that are only warnings. These are a list in "warning_only".
		warningListInterface, keyExists := ruleGroup["warning_only"]
		if keyExists {
			warningList, castOk := warningListInterface.([]string)
			if castOk {
				for _, warningEntry := range warningList {
					fullRuleName := fmt.Sprintf("%s__%s", ruleGroupName, warningEntry)
					warningOnly[fullRuleName] = struct{}{}
				}
			}
		}

		// Loop through rules in a group
		for ruleName, ruleConfig := range ruleGroup {
			if ruleName == "warning_only" {
				continue
			}

			fullRuleName := fmt.Sprintf("%s__%s", ruleGroupName, ruleName)

			_, isError := warningOnly[fullRuleName]
			if isError {
				cfg.WarningOnly[fullRuleName] = struct{}{}
			}

			err := cfg.addRuleFromConfig(fullRuleName, ruleConfig)
			if err != nil {
				return fmt.Errorf("rule %s has invalid config: %w", fullRuleName, err)
			}
		}
	}

	return nil
}

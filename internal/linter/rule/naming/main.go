// Package naming contains rules related to naming convention of action and workflow fields.
package naming

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	// ValueDashCase is a configuration value indicating that a field should follow the dash-case convention.
	ValueDashCase = "dash-case"
	// ValueDashCaseUnderscore is a configuration value indicating that a field should follow the dash-case with
	// underscore prefix allowed convention.
	ValueDashCaseUnderscore = "dash-case;underscore-prefix-allowed"
	// ValueCamelCase is a configuration value indicating that a field should follow the camel-case convention.
	ValueCamelCase = "camelCase"
	// ValuePascalCase is a configuration value indicating that a field should follow the pascal-case convention.
	ValuePascalCase = "PascalCase"
	// ValueAllCaps is a configuration value indicating that a field should follow the all-caps-case convention.
	ValueAllCaps = "ALL_CAPS"
)

var (
	errValueNotString = errors.New("value should be string")
	errValueNotValid  = fmt.Errorf(
		"value can be one of: %s, %s, %s, %s",
		ValueDashCase, ValueCamelCase, ValuePascalCase, ValueAllCaps,
	)
	errFileInvalidType = errors.New("file is of invalid type")
)

var regexpRefs = map[string]*regexp.Regexp{
	"env":     regexp.MustCompile("\\${{[ ]*env\\.([a-zA-Z0-9\\-_]+)[ ]*}}"),
	"vars":    regexp.MustCompile("\\${{[ ]*var\\.([a-zA-Z0-9\\-_]+)[ ]*}}"),
	"secrets": regexp.MustCompile("\\${{[ ]*secret\\.([a-zA-Z0-9\\-_]+)[ ]*}}"),
}

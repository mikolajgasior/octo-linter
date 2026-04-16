// Package dependencies contains rules checking various dependencies between action steps, workflow jobs etc.
package dependencies

import (
	"errors"
	"regexp"
)

var (
	errFileInvalidType = errors.New("file is of invalid type")
	errValueNotBool    = errors.New("value should be bool")
)

var (
	regexpRefInput   = regexp.MustCompile(`\${{[ ]*inputs\.([a-zA-Z0-9\-_]+)[ ]*}}`)
	regexpStepOutput = regexp.MustCompile(
		`\${{[ ]*steps\.([a-zA-Z0-9\-_]+)\.outputs\.([a-zA-Z0-9\-_]+)[ ]*}}`,
	)
	regexpAppendToGithubOutput = regexp.MustCompile(
		`echo[ ]+["']([a-zA-Z0-9\-_]+)=.*["'][ ]+.*>>[ ]+["]{0,1}\$GITHUB_OUTPUT["]{0,1}`,
	)
	regexpLocal = regexp.MustCompile(
		`^\.\/\.github\/actions\/([a-z0-9\-]+|[a-z0-9\-]+\/[a-z0-9\-]+)$`,
	)
	regexpExternal = regexp.MustCompile(
		`[a-zA-Z0-9\-\_]+\/[a-zA-Z0-9\-\_]+(\/[a-zA-Z0-9\-\_]){0,1}@[a-zA-Z0-9\.\-\_]+`,
	)
)

var regexpRefs = map[string]*regexp.Regexp{
	"env":     regexp.MustCompile("\\${{[ ]*env\\.([a-zA-Z0-9\\-_]+)[ ]*}}"),
	"vars":    regexp.MustCompile("\\${{[ ]*env\\.([a-zA-Z0-9\\-_]+)[ ]*}}"),
	"secrets": regexp.MustCompile("\\${{[ ]*env\\.([a-zA-Z0-9\\-_]+)[ ]*}}"),
}

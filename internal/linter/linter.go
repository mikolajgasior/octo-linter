package linter

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/mikolajgasior/octo-linter/v2/internal/linter/glitch"
	"github.com/mikolajgasior/octo-linter/v2/internal/linter/rule"
	"github.com/mikolajgasior/octo-linter/v2/pkg/dotgithub"
)

const (
	// HasNoErrorsOrWarnings indicates that linting completed with no errors or warnings.
	HasNoErrorsOrWarnings = iota

	// HasErrors indicates that one or more rules failed and were classified as errors.
	HasErrors

	// HasOnlyWarnings indicates that some rules failed, but they were configured as warnings only.
	HasOnlyWarnings
)

const (
	// MillisecondsTickerCheckingChannelsClosed is the ticker interval (ms) for checking channel closure.
	MillisecondsTickerCheckingChannelsClosed = 500

	// FileModeOutputMarkdown sets the mode for the generated markdown summary file.
	FileModeOutputMarkdown = 0o600
)

// Linter represents a linter with specific configuration.
type Linter struct {
	Config *Config
}

// Lint runs rules on the given DotGithub and returns the result.
// Optionally writes a Markdown summary to an output file.
//
//nolint:gocognit,funlen
func (l *Linter) Lint(dotGithub *dotgithub.DotGithub, output string, outputLimit int) (int, error) {
	if l.Config == nil {
		panic("Config cannot be nil")
	}

	if dotGithub == nil {
		panic("DotGithub cannot be empty")
	}

	summary := newSummary()
	numCPU := runtime.NumCPU()

	chJobs := make(chan Job)
	chWarnings := make(chan glitch.Glitch)
	chErrors := make(chan glitch.Glitch)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(numCPU)

	go func() {
		for _, action := range dotGithub.Actions {
			for ruleIdx, ruleEntry := range l.Config.Rules {
				if ruleEntry.FileType()&rule.DotGithubFileTypeAction == 0 {
					continue
				}

				isError := l.Config.IsError(ruleEntry.ConfigName(rule.DotGithubFileTypeAction))
				chJobs <- Job{
					rule:      ruleEntry,
					file:      action,
					dotGithub: dotGithub,
					isError:   isError,
					value:     l.Config.Values[ruleIdx],
				}

				summary.numJob.Add(1)
			}
		}

		for _, workflow := range dotGithub.Workflows {
			for ruleIdx, ruleEntry := range l.Config.Rules {
				if ruleEntry.FileType()&rule.DotGithubFileTypeWorkflow == 0 {
					continue
				}

				isError := l.Config.IsError(ruleEntry.ConfigName(rule.DotGithubFileTypeWorkflow))
				chJobs <- Job{
					rule:      ruleEntry,
					file:      workflow,
					dotGithub: dotGithub,
					isError:   isError,
					value:     l.Config.Values[ruleIdx],
				}

				summary.numJob.Add(1)
			}
		}

		close(chJobs)
		waitGroup.Done()
	}()

	go func() {
		for {
			job, more := <-chJobs
			if !more {
				close(chWarnings)
				close(chErrors)

				waitGroup.Done()

				return
			}

			compliant, err := job.Run(chWarnings, chErrors)
			if err != nil {
				slog.Error(
					"error running job",
					slog.String("err", err.Error()),
				)
				summary.numError.Add(1)

				continue
			}

			if !compliant {
				if job.isError {
					summary.numError.Add(1)
				} else {
					summary.numWarning.Add(1)
				}
			}

			summary.numProcessed.Add(1)
		}
	}()

	for range numCPU - 2 {
		go func() {
			chWarningsClosed := false
			chErrorsClosed := false

			ticker := time.NewTicker(MillisecondsTickerCheckingChannelsClosed * time.Millisecond)

			for {
				select {
				case glitchInstance, more := <-chWarnings:
					if more {
						slog.Warn(
							glitchInstance.ErrText,
							slog.String("path", glitchInstance.Path),
							slog.String("rule", glitchInstance.RuleName),
						)

						glitchInstance.IsError = false
						summary.addGlitch(&glitchInstance)
					} else {
						chWarningsClosed = true
					}
				case glitchInstance, more := <-chErrors:
					if more {
						slog.Error(
							glitchInstance.ErrText,
							slog.String("path", glitchInstance.Path),
							slog.String("rule", glitchInstance.RuleName),
						)

						glitchInstance.IsError = true
						summary.addGlitch(&glitchInstance)
					} else {
						chErrorsClosed = true
					}
				case <-ticker.C:
					if chWarningsClosed && chErrorsClosed {
						waitGroup.Done()

						return
					}
				}
			}
		}()
	}

	waitGroup.Wait()

	finalStatus := HasNoErrorsOrWarnings

	if summary.numError.Load() > 0 {
		finalStatus = HasErrors
	} else if summary.numWarning.Load() > 0 {
		finalStatus = HasOnlyWarnings
	}

	slog.Debug(
		"summary",
		slog.Int("rules_returning_errors", int(summary.numError.Load())),
		slog.Int("rules_processed", int(summary.numProcessed.Load())),
		slog.Int("glitches", len(summary.glitches)),
	)

	if output != "" {
		outputMd := filepath.Join(output, "output.md")
		slog.Debug(
			"writing markdown output",
			slog.String("path", outputMd),
		)

		_ = os.Remove(outputMd)

		if outputLimit < 0 {
			outputLimit = 0
		}

		md := summary.markdown("octo-linter summary", outputLimit)

		err := os.WriteFile(outputMd, []byte(md), FileModeOutputMarkdown)
		if err != nil {
			return finalStatus, fmt.Errorf("error writing markdown output: %w", err)
		}
	}

	return finalStatus, nil
}

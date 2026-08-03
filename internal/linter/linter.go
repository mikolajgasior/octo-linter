package linter

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"octo-linter/internal/dotgithub"
	"octo-linter/internal/linter/glitch"
	"octo-linter/internal/linter/rule"
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

	chJobs := make(chan Job)
	chWarnings := make(chan glitch.Glitch)
	chErrors := make(chan glitch.Glitch)

	// Worker pool: one goroutine per CPU, at minimum one.
	numWorkers := max(1, runtime.NumCPU()-1)

	var workerWg sync.WaitGroup

	workerWg.Add(numWorkers)

	for range numWorkers {
		go func() {
			defer workerWg.Done()

			for job := range chJobs {
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
	}

	// Close result channels once all workers have finished sending findings.
	go func() {
		workerWg.Wait()
		close(chWarnings)
		close(chErrors)
	}()

	// Single consumer draining both result channels until both are closed.
	var consumerWg sync.WaitGroup

	consumerWg.Add(1)

	go func() {
		defer consumerWg.Done()

		warnCh := chWarnings
		errCh := chErrors

		for warnCh != nil || errCh != nil {
			select {
			case glitchInstance, more := <-warnCh:
				if more {
					slog.Warn(
						glitchInstance.ErrText,
						slog.String("path", glitchInstance.Path),
						slog.String("rule", glitchInstance.RuleName),
					)

					glitchInstance.IsError = false
					summary.addGlitch(&glitchInstance)
				} else {
					warnCh = nil
				}
			case glitchInstance, more := <-errCh:
				if more {
					slog.Error(
						glitchInstance.ErrText,
						slog.String("path", glitchInstance.Path),
						slog.String("rule", glitchInstance.RuleName),
					)

					glitchInstance.IsError = true
					summary.addGlitch(&glitchInstance)
				} else {
					errCh = nil
				}
			}
		}
	}()

	// Produce jobs in the current goroutine.
	for _, action := range dotGithub.Actions {
		if l.Config.Paths != nil && !l.Config.Paths.Check(action.Path) {
			slog.Info("skipping action due to 'paths' configuration", slog.String("path", action.Path))

			continue
		}

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
		if l.Config.Paths != nil && !l.Config.Paths.Check(workflow.Path) {
			slog.Info("skipping workflow due to 'paths' configuration", slog.String("path", workflow.Path))

			continue
		}

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

	close(chJobs)   // signal workers to stop
	consumerWg.Wait() // wait for consumer to drain all findings

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

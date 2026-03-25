// Package dotgithub reads the contents of a .github directory, parsing all actions and workflows into structured data.
package dotgithub

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mikolajgasior/octo-linter/v2/pkg/action"
	"github.com/mikolajgasior/octo-linter/v2/pkg/workflow"
)

// DotGithub represents contents of .github directory.
type DotGithub struct {
	Actions         map[string]*action.Action
	ExternalActions map[string]*action.Action
	Workflows       map[string]*workflow.Workflow
	Vars            map[string]bool
	Secrets         map[string]bool
}

const (
	// NumExternalActionPathParts defines the number of segments in a 'uses' path split by '/'.
	NumExternalActionPathParts = 3
	// NumExternalActionPathPartsNoSubdir defines the number of segments in a 'uses' path split by '/' when the action
	// is not in a subdirectory.
	NumExternalActionPathPartsNoSubdir = 2
)

const (
	regexpExternalAction = `[a-zA-Z0-9\-\_]+\/[a-zA-Z0-9\-\_]+(\/[a-zA-Z0-9\-\_]){0,1}@[a-zA-Z0-9\.\-\_]+`
)

var (
	errExternalActionNotFound  = errors.New("external action was not found")
	errActionHTTPRequestDo     = errors.New("error doing http request for action yaml")
	errActionHTTPRequestCreate = errors.New("error creating http request for action yaml")
)

func errCreatingHTTPRequestForAction(err error) error {
	return fmt.Errorf("%w: %s", errActionHTTPRequestCreate, err.Error())
}

func errDoingHTTPRequestForAction(err error) error {
	return fmt.Errorf("%w: %s", errActionHTTPRequestDo, err.Error())
}

// ReadDir scans the given directory and parses all GitHub Actions workflow and action YAML files into the struct.
func (d *DotGithub) ReadDir(ctx context.Context, path string) error {
	d.Actions = make(map[string]*action.Action)
	d.Workflows = make(map[string]*workflow.Workflow)

	err := d.getActionsFromDir(path)
	if err != nil {
		return fmt.Errorf("error getting actions from dir %s: %w", path, err)
	}

	err = d.getWorkflowsFromDir(path)
	if err != nil {
		return fmt.Errorf("error getting workflows from dir %s: %w", path, err)
	}

	err = d.processActions(ctx)
	if err != nil {
		return fmt.Errorf("error processing struct actions: %w", err)
	}

	err = d.processWorkflows(ctx)
	if err != nil {
		return fmt.Errorf("error processing struct workflows: %w", err)
	}

	return nil
}

// ReadVars reads a file with GitHub Actions variables, parsing each line into the struct as a variable.
func (d *DotGithub) ReadVars(path string) error {
	if path == "" {
		return nil
	}

	d.Vars = make(map[string]bool)

	slog.Debug(
		"reading file with list of possible variable names...",
		slog.String("path", path),
	)

	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading vars file %s: %w", path, err)
	}

	lines := strings.Fields(string(b))
	for _, variable := range lines {
		d.Vars[variable] = true
	}

	return nil
}

// ReadSecrets reads a file with GitHub Actions secrets, parsing each line into the struct as a secret.
func (d *DotGithub) ReadSecrets(path string) error {
	if path == "" {
		return nil
	}

	d.Secrets = make(map[string]bool)

	slog.Debug(
		"reading file with list of possible secret names...",
		slog.String("path", path),
	)

	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error reading secrets file %s: %w", path, err)
	}

	lines := strings.Fields(string(b))
	for _, secret := range lines {
		d.Secrets[secret] = true
	}

	return nil
}

// GetAction returns an Action by its name.
func (d *DotGithub) GetAction(name string) *action.Action {
	return d.Actions[name]
}

// GetExternalAction returns an Action that is defined outside the current repository, by name.
func (d *DotGithub) GetExternalAction(name string) *action.Action {
	if d.ExternalActions == nil {
		d.ExternalActions = map[string]*action.Action{}
	}

	return d.ExternalActions[name]
}

// DownloadExternalAction downloads a GitHub Action from its “uses” path (e.g., "actions/checkout@v4").
func (d *DotGithub) DownloadExternalAction(ctx context.Context, path string) error {
	if d.ExternalActions == nil {
		d.ExternalActions = map[string]*action.Action{}
	}

	if d.ExternalActions[path] != nil {
		return nil
	}

	repoVersion := strings.Split(path, "@")
	ownerRepoDir := strings.SplitN(repoVersion[0], "/", NumExternalActionPathParts)

	directory := ""
	if len(ownerRepoDir) > NumExternalActionPathPartsNoSubdir {
		directory = "/" + ownerRepoDir[2]
	}

	actionURLPrefix := fmt.Sprintf(
		"https://raw.githubusercontent.com/%s/%s/%s%s",
		ownerRepoDir[0],
		ownerRepoDir[1],
		repoVersion[1],
		directory,
	)

	resp, err := d.getActionHTTPResponse(ctx, actionURLPrefix)
	if err != nil {
		return fmt.Errorf("error getting response from http request to action: %w", err)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Error(
				"error closing response body",
				slog.String("err", err.Error()),
			)
		}
	}()

	b, _ := io.ReadAll(resp.Body)

	d.ExternalActions[path] = &action.Action{
		Path:    path,
		DirName: "",
		Raw:     b,
	}

	err = d.ExternalActions[path].Unmarshal(true)
	if err != nil {
		return fmt.Errorf("error unmarshaling external action: %w", err)
	}

	return nil
}

// IsVarExist checks whether the variable has been loaded from the variables file.
func (d *DotGithub) IsVarExist(name string) bool {
	_, ok := d.Vars[name]

	return ok
}

// IsSecretExist checks whether the secret has been loaded from the secrets file.
func (d *DotGithub) IsSecretExist(name string) bool {
	_, ok := d.Secrets[name]

	return ok
}

func (d *DotGithub) getActionsFromDir(path string) error {
	dirActions := filepath.Join(path, "actions")

	entries, err := os.ReadDir(dirActions)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error reading actions directory: %w", err)
		}
	}

	for _, entry := range entries {
		dirAction := filepath.Join(dirActions, entry.Name())

		// only directories
		fileInfo, err := os.Stat(dirAction)
		if err != nil {
			return fmt.Errorf("error getting os.Stat on %s: %w", dirAction, err)
		}

		if !fileInfo.IsDir() {
			continue
		}

		// search for action.yml or action.yaml file
		ymlAction := filepath.Join(dirAction, "action.yml")
		_, err = os.Stat(ymlAction)

		ymlNotFound := os.IsNotExist(err)
		if err != nil && !ymlNotFound {
			return fmt.Errorf("error getting os.Stat on %s: %w", ymlAction, err)
		}

		if ymlNotFound {
			yamlAction := filepath.Join(dirAction, "action.yaml")
			_, err = os.Stat(yamlAction)

			yamlNotFound := os.IsNotExist(err)
			if err != nil && !yamlNotFound {
				return fmt.Errorf("error getting os.Stat on %s: %w", yamlAction, err)
			}

			if !yamlNotFound {
				ymlAction = yamlAction
			} else {
				continue
			}
		}

		d.Actions[entry.Name()] = &action.Action{
			Path:    ymlAction,
			DirName: entry.Name(),
		}
	}

	return nil
}

func (d *DotGithub) getWorkflowsFromDir(path string) error {
	dirWorkflows := filepath.Join(path, "workflows")

	entries, err := os.ReadDir(dirWorkflows)
	if err != nil {
		return fmt.Errorf("error reading workflows directory %s: %w", dirWorkflows, err)
	}

	nameRegex := regexp.MustCompile(`\.y[a]{0,1}ml$`)
	for _, entry := range entries {
		m := nameRegex.MatchString(entry.Name())
		if !m {
			continue
		}

		ymlWorkflow := filepath.Join(dirWorkflows, entry.Name())

		fileInfo, err := os.Stat(ymlWorkflow)
		if err != nil {
			return fmt.Errorf("error getting os.Stat on %s: %w", ymlWorkflow, err)
		}

		if !fileInfo.Mode().IsRegular() {
			continue
		}

		d.Workflows[entry.Name()] = &workflow.Workflow{
			Path: ymlWorkflow,
		}
	}

	return nil
}

func (d *DotGithub) processActions(ctx context.Context) error {
	// download all external actions used in actions' steps
	reExternal := regexp.MustCompile(regexpExternalAction)

	for _, action := range d.Actions {
		err := action.Unmarshal(false)
		if err != nil {
			return fmt.Errorf("error unmarshaling action: %w", err)
		}

		if action.Runs == nil || len(action.Runs.Steps) == 0 {
			continue
		}

		for stepIdx, step := range action.Runs.Steps {
			if !reExternal.MatchString(step.Uses) {
				continue
			}

			err := d.DownloadExternalAction(ctx, step.Uses)
			if err != nil {
				slog.Error(
					"error downloading external action",
					slog.String("action", action.DirName),
					slog.Int("step", stepIdx),
					slog.String("uses", step.Uses),
					slog.String("err", err.Error()),
				)
			}
		}
	}

	return nil
}

func (d *DotGithub) processWorkflows(ctx context.Context) error {
	// download all external actions used in actions' steps
	reExternal := regexp.MustCompile(regexpExternalAction)

	for _, workflow := range d.Workflows {
		err := workflow.Unmarshal(false)
		if err != nil {
			return fmt.Errorf("error unmarshaling workflow: %w", err)
		}

		if len(workflow.Jobs) == 0 {
			continue
		}

		for _, job := range workflow.Jobs {
			if len(job.Steps) == 0 {
				continue
			}

			for stepIdx, step := range job.Steps {
				if !reExternal.MatchString(step.Uses) {
					continue
				}

				err := d.DownloadExternalAction(ctx, step.Uses)
				if err != nil {
					slog.Error(
						"error downloading external action",
						slog.String("workflow", workflow.FileName),
						slog.Int("step", stepIdx),
						slog.String("uses", step.Uses),
						slog.String("err", err.Error()),
					)
				}
			}
		}
	}

	return nil
}

func (d *DotGithub) getActionHTTPResponse(
	ctx context.Context,
	urlPrefix string,
) (*http.Response, error) {
	urlYML := urlPrefix + "/action.yml"
	slog.Debug(
		"downloading external action yaml",
		slog.String("url", urlYML),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		urlYML,
		strings.NewReader(""),
	)
	if err != nil {
		return nil, errCreatingHTTPRequestForAction(err)
	}

	httpClient := &http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errDoingHTTPRequestForAction(err)
	}

	if resp.StatusCode != http.StatusOK {
		urlYAML := urlPrefix + "/action.yaml"
		slog.Debug(
			"downloading external action yaml",
			slog.String("url", urlYAML),
		)

		req, err = http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			urlYAML,
			strings.NewReader(""),
		)
		if err != nil {
			return nil, errCreatingHTTPRequestForAction(err)
		}

		resp, err = httpClient.Do(req)
		if err != nil {
			return nil, errDoingHTTPRequestForAction(err)
		}

		if resp.StatusCode != http.StatusOK {
			return nil, errExternalActionNotFound
		}
	}

	return resp, nil
}

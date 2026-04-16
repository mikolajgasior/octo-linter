// Package main contains utils.
package main

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

// S represents a struct that is about to be generated.
type S struct {
	N string
	F map[string]string
}

const (
	// FileModeConfigRules defines mode for the config rules generated file.
	FileModeConfigRules = 0o600
)

func main() {
	genPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if len(os.Args) > 1 && os.Args[1] == "../../" {
		genPath = filepath.Join(genPath, os.Args[1])
	}

	tplObj := struct {
		Rules map[string]S
	}{
		Rules: map[string]S{
			"filenames__action_filename_extensions_allowed": {
				N: "filenames.FilenameExtensionsAllowed",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"filenames__action_directory_name_format": {
				N: "filenames.ActionDirectoryNameFormat",
			},
			"filenames__workflow_filename_extensions_allowed": {
				N: "filenames.FilenameExtensionsAllowed",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"filenames__workflow_filename_base_format": {
				N: "filenames.WorkflowFilenameBaseFormat",
			},
			"workflow_runners__not_latest": {
				N: "runners.NotLatest",
			},
			"referenced_variables_in_actions__not_one_word": {
				N: "refvars.NotOneWord",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"referenced_variables_in_actions__not_in_double_quotes": {
				N: "refvars.NotInDoubleQuotes",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"referenced_variables_in_workflows__not_one_word": {
				N: "refvars.NotOneWord",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"referenced_variables_in_workflows__not_in_double_quotes": {
				N: "refvars.NotInDoubleQuotes",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"dependencies__workflow_needs_field_must_contain_already_existing_jobs": {
				N: "dependencies.WorkflowNeedsWithExistingJobs",
			},
			"dependencies__action_referenced_input_must_exists": {
				N: "dependencies.ReferencedInputExists",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"dependencies__action_referenced_step_output_must_exist": {
				N: "dependencies.ActionReferencedStepOutputExists",
			},
			"dependencies__workflow_referenced_variable_must_exists_in_attached_file": {
				N: "dependencies.WorkflowReferencedVariableExistsInFile",
			},
			"dependencies__workflow_referenced_input_must_exists": {
				N: "dependencies.ReferencedInputExists",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"used_actions_in_action_steps__source": {
				N: "usedactions.Source",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"used_actions_in_action_steps__must_exist": {
				N: "usedactions.Exists",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"used_actions_in_action_steps__must_have_valid_inputs": {
				N: "usedactions.ValidInputs",
				F: map[string]string{"FileTypeRequired": `"action"`},
			},
			"used_actions_in_workflow_job_steps__source": {
				N: "usedactions.Source",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"used_actions_in_workflow_job_steps__must_exist": {
				N: "usedactions.Exists",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"used_actions_in_workflow_job_steps__must_have_valid_inputs": {
				N: "usedactions.ValidInputs",
				F: map[string]string{"FileTypeRequired": `"workflow"`},
			},
			"naming_conventions__action_input_name_format": {
				N: "naming.Action",
				F: map[string]string{"Field": `naming.ActionFieldInputName`},
			},
			"naming_conventions__action_output_name_format": {
				N: "naming.Action",
				F: map[string]string{"Field": `naming.ActionFieldOutputName`},
			},
			"naming_conventions__action_referenced_variable_format": {
				N: "naming.Action",
				F: map[string]string{"Field": `naming.ActionFieldReferencedVariable`},
			},
			"naming_conventions__action_step_env_format": {
				N: "naming.Action",
				F: map[string]string{"Field": `naming.ActionFieldStepEnv`},
			},
			"naming_conventions__workflow_env_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldEnv`},
			},
			"naming_conventions__workflow_job_env_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldJobEnv`},
			},
			"naming_conventions__workflow_job_step_env_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldJobStepEnv`},
			},
			"naming_conventions__workflow_referenced_variable_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldReferencedVariable`},
			},
			"naming_conventions__workflow_dispatch_input_name_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldDispatchInputName`},
			},
			"naming_conventions__workflow_call_input_name_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldCallInputName`},
			},
			"naming_conventions__workflow_job_name_format": {
				N: "naming.Workflow",
				F: map[string]string{"Field": `naming.WorkflowFieldJobName`},
			},
			"naming_conventions__workflow_single_job_only_name": {
				N: "naming.WorkflowSingleJobOnlyName",
			},
			"required_fields__action_requires": {
				N: "required.Action",
				F: map[string]string{"Field": `required.ActionFieldAction`},
			},
			"required_fields__action_input_requires": {
				N: "required.Action",
				F: map[string]string{"Field": `required.ActionFieldInput`},
			},
			"required_fields__action_output_requires": {
				N: "required.Action",
				F: map[string]string{"Field": `required.ActionFieldOutput`},
			},
			"required_fields__workflow_requires": {
				N: "required.Workflow",
				F: map[string]string{"Field": `required.WorkflowFieldWorkflow`},
			},
			"required_fields__workflow_dispatch_input_requires": {
				N: "required.Workflow",
				F: map[string]string{"Field": `required.WorkflowFieldDispatchInput`},
			},
			"required_fields__workflow_call_input_requires": {
				N: "required.Workflow",
				F: map[string]string{"Field": `required.WorkflowFieldCallInput`},
			},
			"required_fields__workflow_requires_uses_or_runs_on_required": {
				N: "required.WorkflowUsesOrRunsOn",
			},
		},
	}

	tpl, err := os.ReadFile(
		filepath.Join(filepath.Clean(genPath), "internal", "linter", "config_rules.go.tpl"),
	)
	if err != nil {
		panic("error opening template file: " + err.Error())
	}

	fileRules, err := os.OpenFile(
		filepath.Join(filepath.Clean(genPath), "internal", "linter", "generated_config_rules.go"),
		os.O_RDWR|os.O_CREATE,
		FileModeConfigRules,
	)
	if err != nil {
		panic("error opening file to write to: " + err.Error())
	}

	defer func() {
		err := fileRules.Close()
		if err != nil {
			log.Printf("error closing file that was written: %s", err.Error())
		}
	}()

	buf := &bytes.Buffer{}
	t := template.Must(template.New("gend_tpl").Parse(string(tpl)))

	err = t.Execute(buf, &tplObj)
	if err != nil {
		panic("error executing template: " + err.Error())
	}

	_, err = fileRules.Write(buf.Bytes())
	if err != nil {
		panic("error writing generated template: " + err.Error())
	}

	log.Printf(
		"Generated %s",
		filepath.Join(genPath, "internal", "linter", "generated_config_rules.go"),
	)
}

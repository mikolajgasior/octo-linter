# Configuration

octo-linter requires a configuration file that defines compliance rules. This section explains how to create one in more detail.

## Initialise configuration file
Use `init` command that will create a sample configuration file named `dotgithub.yml` in
current directory. Use `-d` to write it in another place.

## Requirements
Let’s consider a GitHub repository that contains workflows and actions within the `.github` directory. Several 
developers are contributing to it, and we want to enforce the following rules for the files in that directory:

* Action names must be in a dash-case format
* Action and workflow files should have a .yml extension
* Named-value variables should not be enclosed in double quotes
* The use of the latest runner version should be avoided
* Actions, along with their inputs and outputs (where applicable), must include both `name` and `description` fields
* Only local actions should be used
* Environment variable names in steps must be ALL_CAPS

Additionally, it would be useful to automatically verify that all referenced inputs, outputs, and similar entities are properly defined.

There are many more possible rules, but we’ll focus on these for the purpose of this example.

## Configuration file
Tweak the configuration file with rules that the application would use.

Based on the list in previous section, the configuration file can look as shown below.

````yaml
version: '3'
rules:
  filenames:
    action_filename_extensions_allowed: ['yml'] # Action files should have a .yml extension
    action_directory_name_format: dash-case # Action names must be in a dash-case format
    workflow_filename_extensions_allowed: ['yml'] # Workflow files should have a .yml extension
    warning_only:
      - action_directory_name
      - action_filename_extension
      - workflow_filename_extension

  naming_conventions:
    action_step_env_format: ALL_CAPS # Environment variable names in steps must be ALL_CAPS

  action_required_fields: # Actions, along with their inputs and outputs (where applicable), must include both name and description fields
    action_requires: ['name', 'description']
    action_input_requires: ['description']
    action_output_requires: ['description']

  referenced_variables_in_actions:
    not_in_double_quote: true # Named-value variables should not be enclosed in double quotes

  used_actions_in_action_steps: # Only local actions should be used
    source: local-only
  
  used_actions_in_workflow_job_steps:
    source: local-only

  dependencies: # Verify that all referenced inputs, outputs, and similar entities are properly defined
    action_referenced_input_must_exists: true 
    action_referenced_step_output_must_exist: true
    workflow_referenced_input_must_exists: true
    workflow_referenced_variable_must_exists_in_attached_file: true

  workflow_runners:
    not_latest: true # The use of the latest runner version should be avoided
````

### Warning instead of an error
A non-compliant rule can be treated either as an error or a warning. If a rule is intended to trigger only a warning, it should be included in the `warning_only` list, as shown on above example under the `filenames` rule group.

### Override external action
When a GitHub action that is private is used, octo-linter will not be able to download it. In such cases, it is possible to override the action with a local copy.
To do so, add the action to the `overrides.external_actions_paths` list. See an example below.

````yaml
overrides:
  external_actions_paths:
    my-organisation/repo-name@v1: ../repo-name
````

### Regular expressions for action output checks 
Actions outputs are not always defined, and sometimes they do not have a fixed name. The name can depend on the input value. One of the examples of such actions is 
[aws-actions/amazon-ecr-login](https://github.com/aws-actions/amazon-ecr-login). It creates outputs such as `docker_username_123456789012_dkr_ecr_aws_region_1_amazonaws_com`.
Such output cannot be defined in the configuration file, and therefore it is not possible to check it. To overcome this limitation, octo-linter allows to use 
regular expressions to match the output name.
See an example below.

````yaml
overrides:
  external_actions_outputs:
    aws-actions/amazon-ecr-login:
      - ^docker_(username|password)_[0-9]+_dkr_ecr_aws_[a-z1-6_]+_amazonaws_com$
      - ^docker_(username|password)_public_ecr_aws$
````

With the above configuration, the `dependencies.action_referenced_step_output_must_exist` rule will try to match the output name against the regular expressions.

### Version compatibility
The latest `v2` version of the application supports only configuration version `'3'`. Older configuration versions are no longer supported and would 
require using the previous `v1` release of octo-linter.

Continue to the next section to learn how to run `octo-linter` using the prepared configuration.

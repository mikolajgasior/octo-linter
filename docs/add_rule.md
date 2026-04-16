# New rule

## Requirements

To add a new rule, the following items must be implemented:

* A rule struct that implements the `rule.Rule` interface, including methods such as `Validate` and `Lint`.
* The rule struct should be placed in a new or existing rule group. See directories under `internal/linter/rule`.
* A default configuration entry must be added to `internal/linter/dotgithub.yml`.
* The rule must be linked to its configuration key in the `gen.go` file.
* Tests must be written.
* Documentation must be updated.

See the sections below for more detailed guidance.

:warning: **A single rule struct can serve multiple configuration keys. Please read the full documentation for details.**

### Rule group

Each rule belongs to a group. In the configuration file, these appear as second-level keys under the `rules` section — for example,
`naming_conventions` or `required_fields`.

These groups map to directories under `internal/rules`. For example:
* `naming_conventions` → `internal/rules/naming`
* `required_fields` → `internal/rules/required`

You can place your new rule in an existing group or create a new one.

### Rule struct

The simplest way to start is by copying an existing rule and modifying it.

Every rule must implement the `Rule` interface from `internal/linter/rule`, shown below:
```go
type Rule interface {
	Validate(conf interface{}) error
	Lint(
		config interface{},
		f dotgithub.File,
		d *dotgithub.DotGithub,
		chErrors chan<- glitch.Glitch,
	) (bool, error)
	ConfigName(fileType int) string
	FileType() int
}
```

* `Validate`: Checks if the configuration value is valid.
* `Lint`: Runs the lint logic against a given file (workflow or action) using the provided configuration.
* `ConfigName`: Returns the configuration key associated with the rule. The method receives an integer indicating the file type (action or workflow).
* `FileType`: Returns an integer bitmask indicating which file types this rule applies to.

#### FileType method
If the rule applies only to action files:
```go
func (r ActionReferencedStepOutputExists) FileType() int {
	return rule.DotGithubFileTypeAction
}
```

If the rule applies to both action and workflow files:
```go
func (r ActionReferencedStepOutputExists) FileType() int {
	return rule.DotGithubFileTypeAction | rule.DotGithubFileTypeWorkflow
}
```

To prevent the rule from running twice, it should have a unique field like `FileTypeRequired` which indicates whether the rule will be used for action or workflow.
```go
type NotInDoubleQuotes struct {
	FileTypeRequired string
}
```

In addition, the `Lint` method must have a check for `FileTypeRequired` to ensure the rule is only run for the appropriate file type.
```go
func (r NotInDoubleQuotes) Lint(conf interface{}, file dotgithub.File, _ *dotgithub.DotGithub, chErrors chan<- glitch.Glitch) (bool, error) {
    // ...
    var fileTypeRequired int
    if r.FileTypeRequired == "action" {
        fileTypeRequired = rule.DotGithubFileTypeAction
    }
    if r.FileTypeRequired == "workflow" {
        fileTypeRequired = rule.DotGithubFileTypeWorkflow
    }

    if file.GetType() != fileTypeRequired {
        return true, nil
    }
    // ...
}

```

#### ConfigName method
This method can vary depending on whether the rule struct handles:

A single rule
```go
func (r ActionReferencedStepOutputExists) ConfigName(int) string {
	return "dependencies__action_referenced_step_output_must_exist"
}
```

Different keys for different file types
```go
func (r ReferencedInputExists) ConfigName(t int) string {
	switch t {
	case rule.DotGithubFileTypeWorkflow:
		return "dependencies__workflow_referenced_input_must_exists"
	case rule.DotGithubFileTypeAction:
		return "dependencies__action_referenced_input_must_exists"
	default:
		return "dependencies__*_referenced_input_must_exists"
	}
}
```

Keys based on a custom field
```go
func (r Action) ConfigName(int) string {
	switch r.Field {
	case ActionFieldAction:
		return "required_fields__action_requires"
	case ActionFieldInput:
		return "required_fields__action_input_requires"
	case ActionFieldOutput:
		return "required_fields__action_output_requires"
	default:
		return "required_fields__action_*_requires"
	}
}
```

#### Lint method
To distinguish linting issues from internal errors, use `glitch.Glitch` instances and send them to the `chErrors` channel.

Use existing rules as a reference. Locate a similar rule in the configuration file (`internal/linter/dotgithub.yml`) and review its implementation.

#### Validate method
Use existing rules as templates depending on the type and complexity of the configuration value.

### Configuration file
Your rule must be added to the default configuration file: `internal/linter/dotgithub.yml`. This defines default values and enables the rule by default.

### Link configuration key with rule struct
When octo-linter parses the configuration file, it must map each configuration key to a rule struct. This is done using the registry generated in `gen.go`.

Refer back to the three `ConfigName` method patterns. Below are the corresponding `gen.go` entries:

Single Rule
```go
			"dependencies__action_referenced_step_output_must_exist": {
				N: "dependencies.ActionReferencedStepOutputExists",
			},
```

Multiple Keys for File Types
```go

			"dependencies__action_referenced_input_must_exists": {
				N: "dependencies.ReferencedInputExists",
			},
			// ...
			"dependencies__workflow_referenced_input_must_exists": {
				N: "dependencies.ReferencedInputExists",
			},
```

Rule Struct with Custom Field
```go
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
```

### Documentation
Once your rule is implemented and tested, don’t forget to document it thoroughly. This ensures others understand its purpose and usage.

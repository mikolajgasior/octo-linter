# Running locally

## Syntax
Check below help message for `lint` command:

    Usage:  octo-linter lint [FLAGS]

    Runs the linter on files from a specific directory
    
    Required flags:
    -p,		 --path DIR 		Path to .github directory
    
    Optional flags:
    -c,		 --config FILE 			Linter config with rules in YAML format
    -l,		 --loglevel  			One of INFO,ERR,WARN,DEBUG
    -m,		 --logmultiline  		Each log entry key in a separate line
    -o,		 --output DIR 			Path to where summary markdown gets generated
    -u,		 --output-errors INT 	Limit numbers of errors shown in the markdown output file
    -s,		 --secrets-file  		Check if secret names exist in this file (one per line)
    -z,		 --vars-file  			Check if variable names exist in this file (one per line)

Use `-p` argument to point to `.github` directories.  The tool will search for any actions in the `actions`
directory, where each action is in its own sub-directory and its filename is either `action.yaml` or
`action.yml`.  And, it will search for workflows' `*.yml` and `*.yaml` files in `workflows` directory.

Additionally, all the variable names (meaning `${{ var.NAME }}`) as well as secrets (`${{ secret.NAME }}`)
in the workflow can be checked against a list of possible names.  Use `-z` and `-s` arguments with paths
to files containing a list of possible variable or secret names, with names being separated by new line or
space.  Check [Demo](demo.md) for a sample usage.

## Download
If not compiled, binary can be download from [repository releases](https://github.com/mikolajgasior/octo-linter/releases).

## Using binary
Tweak below command with a path pointing to `.github` and configuration file:

````
./octo-linter lint -p /path/to/.github -l WARN -c config.yaml -m
````

## Using docker image
````
docker run --rm --name octo-linter \
  -v /path/to/.github:/dot-github -v $(pwd):/config \
  mikolajgasior/octo-linter:v2.6.1 \
  lint -p /dot-github -l WARN -c /config/config.yml -m
````

## Checking secrets and vars
Check [Demo](demo.md) page to see an example with checking called `secrets` and `vars`.

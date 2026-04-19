# GitHub Actions Workflow

Below is an example of a workflow that uses octo-linter docker to check files in `.github`.

````yaml
---
name: GitHub Actions YAML linter

on:
  pull_request:
    paths:
      - '.github/**.yml'
      - '.github/**.yaml'

jobs:
  main:
    name: Lint
    runs-on: ubuntu-24.04
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run octo-linter
        run: |
          # we assume that the dotgithub.yml file 
          # is present in .github directory
          docker run --rm --name octo-linter \
            -v $(pwd)/.github:/dot-github \
            mikolajgasior/octo-linter:v2.6.1 \
            lint -p /dot-github -l WARN -m
````

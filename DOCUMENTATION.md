# Terrplate Documentation

This page contains the initial workings of the Terraplate documentation.

## Installation

Terraplate is written in Go and uses the awesome [GoReleaser](https://goreleaser.com/) project to release and publish.

### Docker

```bash
docker pull verifa/terraplate
```

### Brew

```bash
brew install verifa/tap/terraplate
```

### Standalone

Check the GitHub Releases: <https://github.com/verifa/terraplate/releases>

## How it works

Terraplate works by traversing up and down from the working directory detecting Terraplate files ("Terrafiles" for short).
Terrafiles are detected by either being called `terraplate.hcl` or with the suffix `.tp.hcl`.

Terraplate builds a tree of Terrafiles (based on the directory hierarchy), with leaf nodes representing Terraform [Root Modules](https://www.terraform.io/language/modules#the-root-module) (i.e. Terraform should be invoked from this directory).
The Terrafiles are inherited and merged, so any configurations provided in an ancestor Terrafile will be inherited by descendant Terrafiles.

Alongside Terrafiles you can place a `templates` directory which contains files that should be templated to all nested Terrafiles.
Templates with the suffix `.tf` are automatically detected, and a good practice is to name files with a suffix `.tp.tf` so that when the templates are built it is easy to identify the files that came from Terraplate.

Terraplate also generates a `terraplate.tf` file containing things like variables with default values, the `terraform {}` block with the `required_version` and `required_providers` (if specified).

## Terrafile Config

Here are the different configuration blocks that are supported.

### variables

`variables` block defines a map of Terraform variables that will be written to the `terraplate.tf` file with a default.

Use the `variables` block for controlling variables in your root modules, prime examples being things like `environment`, `region`, `project`, etc.

Example:

```hcl
// env.tp.hcl
variables {
  environment = "dev"
}
```

Output:

```hcl
// terraplate.tf
variable "environment" {
  default = "dev"
}
```

Use this in your Terraform files as a normal Terraform variable, e.g. `${var.environment}`

### values

`values` block defines a map of values that are passed to the Go template executor when running the Terraplate build process.

Use this instead of `variables` if you do not want to expose values as Terraform variables but only want to use them during the build process.
A prime example of this, is for configuring the Terraform backend because variables cannot be used for this.

Example:

```hcl
// terraplate.hcl
values {
    some_value = "hello!"
}
```

```hcl
// templates/some_value.tp.tf
locals {
    some_value = "{{ .Values.some_value }}"
}
```

Output:

```hcl
// some_value.tp.tf
locals {
    some_value = "hello!"
}
```

### template

`template` block defines a template, or overrides a template already detected in the `templates` folder next to a terrafile.

By default, templates will be built and this can be disabled.
This is (maybe?) useful if you want to define all your Terraform files as templates in a top-level `templates` directory and control for which leaf Terrafiles the templates should get built.

Templates can also define non-Terraform files in case you want to just do some general-purpose templating, such as creating Makefiles or something spicy.

Example:

```hcl
// templates/ignore.tp.tf

# We don't want this to get generated, by default

# We can still use the Go templating here, e.g. {{ .Values.some_value }}

```

```hcl
// terraplate.hcl
# Reference the template defined above
template "ignore" {
    # Set the build as false, which is inherited for all subdirectories
    # so that this file is never built, unless a descendent Terrafile
    # sets this back to true
    build = false
}
```

### required_providers

`required_providers` defines the required providers for a Terraform root module. It is built into a `terraform {}` block inside a `terraplate.tf` file.

Example:

```hcl
required_providers {
  local = {
    source  = "hashicorp/local"
    version = "2.1.0"
  }
}
```

Output:

```hcl
// terraplate.tf
terraform {
  // ...
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }
}
```

Terraform Docs: <https://www.terraform.io/language/providers/requirements#requiring-providers>

### required_version

`required_version` accepts a string. It is built into a `terraform {}` block inside a `terraplate.tf` file.

Example:

```hcl
required_version = ">= 1.1.0"
```

Output:

```hcl
// terraplate.tf
terraform {
  required_version = ">= 1.1.0"
  // ...
}
```

Terraform Docs: <https://www.terraform.io/language/settings#specifying-a-required-terraform-version>

## Commands

```bash
$ terraplate --help
DRY Terraform using Go Templates.

Terraplate keeps your Terraform DRY.
Create templates that get built using Go Templates to avoid repeating common
Terraform configurations like providers and backend.

Usage:
  terraplate [command]

Available Commands:
  apply       Runs terraform apply on all subdirectories
  build       Build Terraform files based your Terrafiles
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Runs terraform init on all subdirectories
  plan        Runs terraform plan on all subdirectories
```

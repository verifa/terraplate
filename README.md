# Terraplate - DRY Terraform with Go Templates

[![Go Report Card](https://goreportcard.com/badge/github.com/verifa/terraplate)](https://goreportcard.com/report/github.com/verifa/terraplate)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Terraplate is a very simple wrapper around Terraform aimed at reducing the amount of duplicate code used when working with multiple different Terraform [root modules](https://www.terraform.io/language/modules#the-root-module).

The configurations that are always needed include the backend (or remote) and the use of providers.
Quite often ordinary Terraform code is also duplicated, even when using Terraform modules.
It is the goal of Terraplate to help you reduce the burden of maintaining these by using [Go Templates](https://pkg.go.dev/text/template) to build Terraform code for you.

## Motivation

Terraplate would not exist, the way it does, without [Terragrunt](https://terragrunt.gruntwork.io/).
Being fans of Terragrunt we took a lot from the way Terragrunt works to make Terraplate have a somewhat familiar feeling.

The main reasons for developing Terraplate when Terragrunt exists is the following:

1. Terragrunt uses non-native Terraform syntax.
   1. Terragrunt kinda *created* modules, but the Terraform `module {}` block nowadays feels smoother and nicer than a single `terragrunt.hcl` file.
   2. Terragrunt cannot be used with Terraform Cloud (because it's not vanilla syntax)
   3. Onboarding new people to Terragrunt adds an extra layer of complexity because it does some magic behind the scenes
2. Most of the functions, like `find_in_parent_folders()`, feel quite repetitive when in most cases that's what you want
   1. Terraplate's behaviour is built around directory structure and creates a tree of *Terrafiles* (Terraplate files) which inherit everything from their ancestors by default

There's definitely a ton of stuff you probably cannot do with Terraplate that you can with Terragrunt.
Like mentioned, we are Terragrunt fans and were living to find a happy place using Terraform at scale without Terragrunt, and that's why Terraplate was created.

If you are a Terragrunt user and find useful things missing, please raise an issue :)

## Getting Started

Terraplate works by traversing up and down from the working directory detecting Terraplate files ("Terrafiles" for short).
Terrafiles are detected by either being called `terraplate.hcl` or with the suffix `.tp.hcl`.

Terraplate builds a tree of Terrafiles, with leaf nodes being treated as those which should be "built" (i.e. if a Terrafile doesn't have any nested Terrafiles, it is built by default).
The Terrafiles are inherited and merged, so anything written in an ancestor will be inherited by descendant Terrafiles.

Alongside Terrafiles you can place a `templates` directory which contains files that should be templated to all nested Terrafiles, by default, as Templates are also inherited in the tree of Terrafiles.
Templates with the suffix `.tf` are automatically detected, and a good practice is to name files with a suffix `.tp.tf` so that when the templates are built it is easy to identify the files that came from Terraplate.

Terraplate also generates a `terraplate.tf` file containing things like variables with default values, the `terraform {}` block with the `required_version` and `required_providers` (if specified).

### Terrafile Config

Here are the different configuration blocks that are supported.

#### variables

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

#### values

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

#### template

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

#### required_providers

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

#### required_version

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

### Commands

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

## Project Status

This project is currently in **alpha**.

It is being used in production, but the functionality may change and backwards compatability cannot be guaranteed at this time.

Terraplate does not have a cache, cleanup functionality or overwrite protection.
It's quite dumb, by design, so please be careful that you do not overwrite existing files (use Git, wink wink) and name your template files with a suffix such as `.tp.tf` to add another layer of "protection".

## Examples

See the [examples](./examples) folder.

## License

This code is released under the [Apache-2.0 License](./LICENSE).

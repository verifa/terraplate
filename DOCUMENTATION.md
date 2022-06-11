# Terraplate Documentation

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

Terraplate builds a tree of Terrafiles (based on the directory hierarchy), with leaf nodes representing Terraform [Root Modules](https://www.terraform.io/language/modules#the-root-module) (i.e. Terraform should be invoked from these directories).
The Terrafiles are inherited and merged, so any configurations provided in an ancestor Terrafile will be inherited by descendant Terrafiles.

Alongside Terrafiles you can place a `templates` directory which contains Go template files that can be templated into vanilla Terraform files.
Using the `templates` directory is optional, but considered a practice to keep the templates separate.

Terraplate also generates a `terraplate.tf` file containing things like variables and locals, as well as the `terraform {}` block containing the `required_version` and `required_providers` (if specified).

## Terrafile Config

Here are the different configuration blocks that are supported.

### locals

`locals` block defines a map of Terraform locals that will be written to the `terraplate.tf` file.

Use the `locals` block for controlling values in your root modules, e.g. `environment`, `region`, `project`.

Use `locals` when you want to reference these values in your Terraform code.
Prefer locals over variables unless you want to override something at runtime; in that case `variables` are your friend.

Example:

```hcl
# env.tp.hcl
locals {
  environment = "dev"
}
```

Output:

```hcl
# terraplate.tf
locals {
  environment = "dev"
}
```

Use this in your Terraform files as a normal Terraform variable, e.g. `${local.environment}`

### variables

`variables` block defines a map of Terraform variables that will be written to the `terraplate.tf` file with default values.

Prefer `locals` over variables if you will not be overriding inputs at runtime.

Example:

```hcl
# env.tp.hcl
variables {
  environment = "dev"
}
```

Output:

```hcl
# terraplate.tf
variable "environment" {
  default = "dev"
}
```

Use this in your Terraform files as a normal Terraform variable, e.g. `${var.environment}`

### values

`values` block defines a map of values that are passed to the Go template executor when running the Terraplate build process.

Use this instead of `locals` or `variables` if you do not want to expose values as Terraform variables but only want to use them during the build process.
A prime example of this is configuring the Terraform backend because variables cannot be used for this.

Example:

```hcl
# terraplate.hcl
values {
    some_value = "hello!"
}
```

```hcl
# templates/some_value.tp.tf
locals {
    some_value = "{{ .Values.some_value }}"
}
```

Output:

```hcl
# some_value.tp.tf
locals {
    some_value = "hello!"
}
```

### template

`template` block defines a template that will be built to all child root modules (as Terrafiles inherit from their parents).

Templates can also define non-Terraform files in case you want to just do some general-purpose templating, such as creating Makefiles or something spicy.

Example:

```hcl
# templates/backend.tmpl

terraform {
  backend "s3" {
    bucket  = "bucket-name"
    key     = "{{ .RelativeDir }}/terraform.tfstate"
    region  = "{{ .Locals.aws_region }}"
    encrypt = true
  }
}
```

```hcl
# terraplate.hcl

# Define a template to be built, reading the template we have defined.
# All child terrafiles will inherit and build this template.
template "backend" {
  # read_template is a custom function that parses up the directory tree,
  # looking for a matching template file
  contents = read_template("backend.tmpl")
  # target is optional, and defaults to the template name with a "tp.tf" suffix
  target = "backend.tp.tf"
  # condition is optional, and defaults to true. It specifies whether to build
  # the template or not and supports Go templating
  # condition = ""
  
}

# Templates can also embed the contents directly
template "embedded" {
  contents = <<EOL
    # Template this
  EOL
}
```

### required_providers

`required_providers` defines the required providers for a Terraform root module.
It is built into a `terraform {}` block inside a `terraplate.tf` file.

Example:

```hcl
terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }
}
```

Output:

```hcl
# terraplate.tf
terraform {
  # ...
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
# terraplate.tf
terraform {
  required_version = ">= 1.1.0"
  # ...
}
```

Terraform Docs: <https://www.terraform.io/language/settings#specifying-a-required-terraform-version>

## Commands

Use the `terraplate --help` option to get the complete list of commands, and also more details on a subcommand, e.g. `terraplate build --help`

### `terraplate parse`

The `parse` command will not actually do anything. It is meant as more of a debug command to see which Terrafiles have been detected and the templates that will be built.

In the typical development process, `parse` would be the first command that you call.

### `terraplate build`

The `build` command parses the Terrafiles in your directory structure and builds the templates and Terraform files.

### `terraplate init`

The `init` commands runs `terraform init` in each root module detected.

### `terraplate plan`

The `plan` commands runs `terraform plan` in each root module detected.

### `terraplate apply`

The `apply` commands runs `terraform apply` in each root module detected.

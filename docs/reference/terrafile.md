---
title: "Terrafile"
description: "Terrafile reference"
---

Here are the different configurations that are supported in a Terrafile.

## Locals

`locals` block defines a map of Terraform locals that will be written to the `terraplate.tf` file.

Use the `locals` block for controlling values in your root modules, e.g. `environment`, `region`, `project`.

Use `locals` when you want to reference these values in your Terraform code.
Prefer locals over variables unless you want to override something at runtime; in that case `variables` are your friend.

Example:

```terraform title="terraplate.hcl"
locals {
  environment = "dev"
}
```

Output:

```terraform title="terraplate.tf"
locals {
  environment = "dev"
}
```

Use this in your Terraform files as a normal Terraform variable, e.g. `${local.environment}`

## Variables

`variables` block defines a map of Terraform variables that will be written to the `terraplate.tf` file with default values.

Prefer `locals` over variables if you will not be overriding inputs at runtime.

Example:

```terraform title="terraplate.hcl"
variables {
  environment = "dev"
}
```

Output:

```terraform title="terraplate.tf"
variable "environment" {
  default = "dev"
}
```

Use this in your Terraform files as a normal Terraform variable, e.g. `${var.environment}`

## Values

`values` block defines a map of values that are passed to the Go template executor when running the Terraplate build process.

Use this instead of `locals` or `variables` if you do not want to expose values as Terraform variables but only want to use them during the build process.
A prime example of this is configuring the Terraform backend because variables cannot be used for this.

Example:

```terraform title="terraplate.hcl"
values {
    some_value = "hello!"
}
```

```terraform title="templates/some_value.tp.tf.hcl"
locals {
    some_value = "{{ .Values.some_value }}"
}
```

Output:

```terraform title="some_value.tp.tf"
locals {
    some_value = "hello!"
}
```

## Templates

`template` block defines a template that will be built to all child root modules (as Terrafiles inherit from their parents).

Templates can also define non-Terraform files in case you want to just do some general-purpose templating, such as creating Makefiles or something spicy.
But we're just gonna do plain ol' DRY Terraform.

Example:

```terraform title="backend.tmpl"
terraform {
  backend "s3" {
    bucket  = "bucket-name"
    key     = "{{ .RelativeDir }}/terraform.tfstate"
    region  = "{{ .Locals.aws_region }}"
    encrypt = true
  }
}
```

```terraform title="terraplate.hcl"
# Define a template to be built, reading the template we have defined.
# All child terrafiles will inherit and build this template.
template "backend" {
  # read_template is a custom function that parses up the directory tree,
  # looking for a matching template file
  contents = read_template("backend.tmpl")
  # target is optional, and defaults to the template name with a "tp.tf" suffix
  # (e.g. "backend.tp.tf" for this template)
  target = "backend.tp.tf"
}

# Templates can also embed the contents directly
template "embedded" {
  contents = <<-EOL
    # Template this
  EOL
}

# Templates can also have conditional logic when to build them, which makes them
# very powerful
template "provider_aws_dev" {
  contents = <<-EOL
  provider "aws" {
    alias      = "dev"
    region     = "eu-west-1"
    access_key = "my-access-key"
    secret_key = "my-secret-key"
  }
  EOL
  # Specify a condition, which if it evaluates to true, will build the template
  # in that root module
  condition = "{{ eq .Locals.environment \"dev\" }}"
}
```

## Required Providers

`required_providers` defines the required providers for a Terraform root module.
It is built into a `terraform {}` block inside a `terraplate.tf` file.

Example:

```terraform title="terraplate.hcl"
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

```terraform  title="terraplate.tf"
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

## Required Version

`required_version` accepts a string. It is built into a `terraform {}` block inside a `terraplate.tf` file.

Example:

```terraform title="terraplate.hcl"
terraform {
  required_version = ">= 1.1.0"
}
```

Output:

```terraform title="terraplate.tf"
terraform {
  required_version = ">= 1.1.0"
  # ...
}
```

Terraform Docs: <https://www.terraform.io/language/settings#specifying-a-required-terraform-version>

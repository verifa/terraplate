# Terraplate - DRY Terraform with Go Templates

[![Go Report Card](https://goreportcard.com/badge/github.com/verifa/terraplate)](https://goreportcard.com/report/github.com/verifa/terraplate)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Terraplate is a thin wrapper around Terraform aimed at reducing the amount of duplicate code used when working with multiple different Terraform [root modules](https://www.terraform.io/language/modules#the-root-module).

The Terraform backend (or remote) and provider declarations are two configurations that are needed for every root module.
Often ordinary Terraform code is also duplicated (such as invoking a module, or reading from a config file), which is not solved by simply using modules.
It is the goal of Terraplate to help reduce the burden of maintaining these by using [Go Templates](https://pkg.go.dev/text/template) to build Terraform code for you.

## Motivation

Terraplate would not exist without [Terragrunt](https://terragrunt.gruntwork.io/).
Being fans of Terragrunt we took a lot from the way Terragrunt works to make Terraplate have a familiar feeling.

The main reasons for developing Terraplate when Terragrunt exists is the following:

1. Terragrunt does not use native Terraform syntax.
   1. Terragrunt cannot be used with Terraform Cloud
   2. The Terraform `module {}` block feels smoother and nicer than a single `terragrunt.hcl` file (note that Terragrunt existed before Terraform modules).
   3. Onboarding new people to Terragrunt adds an extra layer of complexity because it does some magic behind the scenes
2. Most of the functions, like `find_in_parent_folders()`, feel quite repetitive when in most cases that's what you want
   1. Terraplate's behaviour is built around directory structure and creates a tree of *Terrafiles* (Terraplate files) which inherit everything from their ancestors by default

There's definitely a ton of stuff you probably cannot do with Terraplate that you can with Terragrunt.
Like mentioned, we are Terragrunt fans and were living to find a happy place using Terraform at scale without Terragrunt, and that's why Terraplate was created.

If you are a Terragrunt user and find useful things missing, please raise an issue or discussion :)

## How it works

Terraplate works by traversing up and down from the working directory detecting Terraplate files ("Terrafiles" for short).
Terrafiles are detected by either being called `terraplate.hcl` or with the suffix `.tp.hcl`.

Terraplate builds a tree of Terrafiles (based on the directory hierarchy), with leaf nodes representing Terraform [Root Modules](https://www.terraform.io/language/modules#the-root-module) (i.e. Terraform should be invoked from this directory).
The Terrafiles are inherited and merged, so any configurations provided in an ancestor Terrafile will be inherited by descendant Terrafiles.

Alongside Terrafiles you can place a `templates` directory which contains files that should be templated to all nested Terrafiles.
Templates with the suffix `.tf` are automatically detected, and a good practice is to name files with a suffix `.tp.tf` so that when the templates are built it is easy to identify the files that came from Terraplate.

Terraplate also generates a `terraplate.tf` file containing things like variables with default values, the `terraform {}` block with the `required_version` and `required_providers` (if specified).

[![asciicast](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ.svg)](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ)

## Documentation

Please check the [documentation](./DOCUMENTATION.md) for more details like installation and configurations.

## Project Status

This project is currently in **alpha**.

It is being used in production, but the functionality may change and backwards compatability cannot be guaranteed at this time.

Terraplate does not have a cache, cleanup functionality or overwrite protection.
It's quite dumb, by design, so please be careful that you do not overwrite existing files (use Git, wink wink) and name your template files with a suffix such as `.tp.tf` to add another layer of "protection".

## Examples

See the [examples](./examples) folder.

## License

This code is released under the [Apache-2.0 License](./LICENSE).

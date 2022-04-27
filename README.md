<img src="./terraplate-logo.svg" width="500">

> DRY Terraform with Go Templates

[![Go Report Card](https://goreportcard.com/badge/github.com/verifa/terraplate)](https://goreportcard.com/report/github.com/verifa/terraplate)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Terraplate is a thin wrapper around Terraform aimed at reducing the amount of duplicate code used when working with multiple different Terraform [root modules](https://www.terraform.io/language/modules#the-root-module).

## Who is it for

**1. Terragrunt users who want to use Terraform Cloud.**

This was one of our use cases when building Terraplate.
Terraplate should provide a familiar feeling and allow you to build vanilla Terraform.

**2. Terraform users with multiple [Root Modules](https://www.terraform.io/language/modules#the-root-module)**

This is related to 1. above, where users have already solved this with Terragrunt.

Once you start to scale your Terraform usage you will not want to put all of your code into a single root module.
That would mean one state containing all of your resources.

For example, when working with Kubernetes it is also [strongly recommended](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs#stacking-with-managed-kubernetes-cluster-resources) to split the cluster creation from the creation of Kubernetes resources using Terraform.

[Terraform Workspaces](https://www.terraform.io/language/state/workspaces#when-to-use-multiple-workspaces) do not solve this, quoting from their own documentation:

> *Workspaces alone are not a suitable tool for system decomposition, because each subsystem should have its own separate configuration and backend, and will thus have its own distinct set of workspaces.*

Terraplate is designed to keep the configurations and backend DRY.

## What it does

Terraplate traverses up and down from the working directory detecting Terraplate files ("Terrafiles" for short), treating the Terrafiles without nested Terrafiles as [Root Modules](https://www.terraform.io/language/modules#the-root-module).

Terraplate builds Terraform files based on your provided Terraform templates (using the Go Templating engine).
Define your Terraform configs once, and use Go Templates to substitute the values based on the different root modules.

The built files are completely normal Terraform files, that should be checked into Git and can be run either via the `terraform` CLI, [Terraform Cloud](https://www.terraform.io/cloud), or using the `terraplate` CLI.

[![asciicast](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ.svg)](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ)

## Motivation

Terraplate would not exist without [Terragrunt](https://terragrunt.gruntwork.io/).
Being fans of Terragrunt we took a lot from the way Terragrunt works to make Terraplate have a familiar feeling.

The main reasons for developing Terraplate when Terragrunt exists is the following:

1. Terragrunt does not use native Terraform syntax.
   1. Terragrunt cannot be used with Terraform Cloud
   2. Onboarding new people to Terragrunt adds an extra layer of complexity
2. Most of the functions, like `find_in_parent_folders()`, feel quite repetitive when in most cases that's what you want
   1. Terraplate's behaviour is built around directory structure and creates a tree of *Terrafiles* (Terraplate files) which inherit everything from their ancestors by default

There's a lot of things you can do with Terragrunt that you cannot do with Terraplate.
Like mentioned, we are Terragrunt fans and have been trying to find a happy place using *just* Terraform, and that's why Terraplate was created.

If you are a Terragrunt user and find useful things missing, please raise an issue or discussion :)

## Documentation

Please check the [documentation](./DOCUMENTATION.md) for more details like installation and configurations.

## Project Status

This project is currently in **alpha**.

It is being used in production, but the functionality may change and backwards compatibility cannot be guaranteed at this time.

Terraplate does not have a cache, cleanup functionality or overwrite protection.
It's quite dumb, by design, so please be careful that you do not overwrite existing files (use Git, wink wink) and name your template files with a suffix such as `.tp.tf` to add another layer of "protection".

## Examples

See the [examples](./examples) folder.

## License

This code is released under the [Apache-2.0 License](./LICENSE).

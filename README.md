<img src="./terraplate-logo.svg" width="500">

> DRY Terraform with Go Templates

[![Go Report Card](https://goreportcard.com/badge/github.com/verifa/terraplate)](https://goreportcard.com/report/github.com/verifa/terraplate)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Terraplate is a thin wrapper around Terraform aimed at reducing the amount of duplicate code used when working with multiple different Terraform [root modules](https://www.terraform.io/language/modules#the-root-module).

## What it does

Terraplate traverses up and down from the working directory detecting Terraplate files (AKA "Terrafiles"), treating the Terrafiles without child Terrafiles as [Root Modules](https://www.terraform.io/language/modules#the-root-module) (i.e. if a Terrafile does not have any children, it's considered a "root module" where Terraform should be run).

Terraplate builds Terraform files based on your provided Terraform templates (using the Go Templating engine).
Define your Terraform configs once, and use Go Templates to substitute the values based on the different root modules.

The built files are completely normal Terraform files, that should be checked into Git and can be run either via the `terraform` CLI or using the `terraplate` CLI.

[![asciicast](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ.svg)](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ)

## Motivation

Terraplate would not exist without [Terragrunt](https://terragrunt.gruntwork.io/).
Being fans of Terragrunt we took a lot from the way Terragrunt works to make Terraplate have a familiar feeling.

The main reasons for developing Terraplate when Terragrunt exists is the following:

1. Terragrunt does not use native Terraform syntax
2. Terraplate has inheritance built in without being explicit (e.g. functions like `find_in_parent_folders()` don't need to be used)

There's a lot of things you can do with Terragrunt that you cannot do with Terraplate.
Like mentioned, we are Terragrunt fans and have been trying to find a happy place using *just* Terraform, and that's why Terraplate was created.
If you start with Terraplate and find it's not for you; that's ok, there's no lock-in as all the files are just vanilla Terraform.

If you are a Terragrunt user and find useful things missing, please raise an issue or discussion :)

## Who is it for

**1. Terraform users with multiple [Root Modules](https://www.terraform.io/language/modules#the-root-module)**

This is related to 1. above, where users have already solved this with Terragrunt or using Terraform [workspaces](https://www.terraform.io/cli/workspaces).

Once you start to scale your Terraform usage you will not want to put all of your code into a single root module.

For example, when working with Kubernetes it is also [strongly recommended](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs#stacking-with-managed-kubernetes-cluster-resources) to split the cluster creation from the creation of Kubernetes resources using Terraform.

**2. Terraform users who want to avoid [Workspaces](https://www.terraform.io/cli/workspaces)**

If you don't find workspaces right for you, Terraplate can avoid lots of copy and paste and provide a better developer experience (avoid having to switch workspaces and instead switch directory). Terraform's own [documentation](https://www.terraform.io/language/state/workspaces#when-to-use-multiple-workspaces) also do not recommend workspaces in certain cases:

> *Workspaces alone are not a suitable tool for system decomposition, because each subsystem should have its own separate configuration and backend, and will thus have its own distinct set of workspaces.*

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

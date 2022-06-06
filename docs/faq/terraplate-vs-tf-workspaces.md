---
title: "Terraplate vs Terraform Workspaces"
description: "Terraplate and Terraform workspaces help keep Terraform DRY. So how do they compare?"
---

## Terraform Workspaces

[Terraform Workspaces](https://www.terraform.io/language/state/workspaces) are a feature of Terraform that enable you to use different backends (state storage) for a single root module (i.e. directory).
This means you can have exactly the same Terraform code, and substitue the variables to manage different environments.
Thus, workspaces help to keep Terraform DRY but only for a *single Root Module* with Terraform variable substitution.

Terraplate, on the other hand, allows you to template any Terraform code (such as backend, providers, resources, module invocations) for any Root Module that has a Terrafile.

Workspaces and Terraplate solve similar issues, but should be not be considered directly as alternatives.
In fact, you can use Workspaces and Terraplate together.

### When to use workspaces

Terraform's own [documentation](https://www.terraform.io/language/state/workspaces#when-to-use-multiple-workspaces) do not recommend workspaces in certain cases:

> *Workspaces alone are not a suitable tool for system decomposition, because each subsystem should have its own separate configuration and backend, and will thus have its own distinct set of workspaces.*

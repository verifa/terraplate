---
title: "Introduction"
description: "Terraplate is a small command line tool to keep Terraform DRY and improve developer productivity when working with Terraform."
lead: "Terraplate is a small command line tool to keep Terraform DRY and improve developer productivity when working with Terraform."
date: 2020-10-06T08:48:57+00:00
lastmod: 2020-10-06T08:48:57+00:00
draft: false
images: []
menu:
  docs:
    parent: "prologue"
weight: 100
toc: true
---

## Getting started

Checkout our tutorials if you are new to Terraplate, or our Quick Start for a one-pager with useful information.

### Tutorial

{{< alert icon="👉" text="The Tutorial is intended for novice to intermediate users." />}}

Step-by-step instructions on how to start a new Doks project. [Tutorial →](https://getdoks.org/tutorial/introduction/)

### Quick Start

{{< alert icon="👉" text="The Quick Start is intended for intermediate to advanced users." />}}

One page summary of how to start a new Doks project. [Quick Start →]({{< relref "quick-start" >}})

## What it does

Terraplate traverses up and down from the working directory detecting Terraplate files (AKA "Terrafiles"), treating the Terrafiles without child Terrafiles as [Root Modules](https://www.terraform.io/language/modules#the-root-module) (i.e. if a Terrafile does not have any children, it's considered a "root module" where Terraform should be run).

Terraplate builds Terraform files based on your provided Terraform templates (using the Go Templates).
Define your Terraform configs once, and use Go Templates to substitute the values based on the different root modules.

The built files are completely normal Terraform files that should be checked into Git and can be run either via the `terraform` CLI or using the `terraplate` CLI.
This way you can focus on writing your Terraform code that creates resources, and let Terraplate handle the boilerplate (like backend, providers, configuration, etc) based on your provided templates.

The goal of Terraplate is to not do any magic: just plain (but DRY) Terraform, which means you can bring your own tools for static analysis, security, policies and testing.

The `terraplate` CLI allows you to run Terraform across all your Root Modules and provide a summary of plans.

> INSERT VIDEO HERE

<!-- [![asciicast](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ.svg)](https://asciinema.org/a/DXAzFxSUWFaYn5iPU8DnliyRZ) -->

## Motivation

As you scale your Terraform usage you will start to split your resources out across multiple Terraform [Root Modules](https://www.terraform.io/language/modules#the-root-module).
Each Root Module must define it's own backend (state storage) and providers that are required, and this can lead to a lot of copy+paste.

There are existing techniques to help alleviate this, two notable mentions:

1. [Terraform Workspaces](https://www.terraform.io/cli/workspaces): only solve the issue when you have multiple environments (e.g. prod & dev) for the same infrastructure, not multiple completely unrelated Root Modules. Nonetheless, it helps reduce the amount of copied code. Check the [FAQ on the subject]({{< relref "../faq/terraplate-vs-tf-workspaces" >}}).
2. [Terragrunt](https://terragrunt.gruntwork.io/): Terraplate would not exist without Terragrunt.Terragrunt inspired Terraplate and therefore it is no surprise that Terraplate has a similar feel. However, there are differences that we feel warranted the development of another tool. Check the [FAQ on the subject]({{< relref "../faq/terraplate-vs-terragrunt" >}}).

## Who is it for

### Terraform users with multiple [Root Modules](https://www.terraform.io/language/modules#the-root-module)

Once you start to scale your Terraform usage you will not want to put all of your code into a single root module (i.e. a single state).

The two main benefits Terraplate brings is:

1. Keeping your code DRY and more maintainable
2. Developer productivity: not just less time writing boilerplate, but also running Terraform across all your Root Modules with a nice summary

### Terraform users who want to make [Workspaces](https://www.terraform.io/cli/workspaces) more DRY or avoid them

If you don't find workspaces completely solves the issue of DRY infra, or they are not right for you, Terraplate is worth considering.
Terraplate is not a replacement, but something that can solve the same problem and be used together with workspaces.
Check the [FAQ on the subject]({{< relref "../faq/terraplate-vs-tf-workspaces" >}}).

### Overcoming limitations of Terraform's dynamic behavior

An example of a limitation is the ability to do `for_each` for providers (or even dynamically reference providers to pass to modules using a `for_each`).
With Terraplate, you can build the `.tf` Terraform file that creates the providers and invokes the modules and overcome this.
It's not the cleanest, but we've found it much friendlier than the numerous workarounds we have to do to achieve the same thing with vanilla Terraform.

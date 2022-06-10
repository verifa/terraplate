---
title: "Terraplate vs Terragrunt"
description: "Terraplate and Terragrunt solve similar problems, in a similar way. So why create Terraplate?"
lead: "Terraplate and Terragrunt solve similar problems in a similar way. So why create Terraplate?"
date: 2020-10-13T15:21:01+02:00
lastmod: 2020-10-13T15:21:01+02:00
draft: false
images: []
menu:
  docs:
    parent: "faq"
weight: 110
toc: true
---

## Native Terraform Syntax

One of the main motivators for *not* using Terragrunt is that it's syntax is not native.
Of course in the end, Terragrunt generates native Terraform code which is invoked using the Terraform CLI, but this happens at runtime and is managed by Terragrunt.

This means:

1. You are dependent on Terraplate to run Terraform
   - This rules out tools like [Terraform Cloud](https://cloud.hashicorp.com/products/terraform)
2. There are subtle differences in syntax between using Terraform features, like modules, which as a concept existed in Terragrunt before they existed in Terraform

## Calling none or multiple modules

Terragrunt is designed to call exactly one Terraform module from each Root Module.
Sometimes you don't need a Terraform module for what you are writing, or you simply haven't gotten far enough yet to create one.
Terragrunt forces you to create Terraform modules, and we found this frustrating.

With Terraplate, you can use the templating engine to create things like the backend and providers, and then write a plain `.tf` file yourself which can do whatever you want it to.

## Keep it DRY, only

Terraplate can be used for *just* it's templating engine.
That way, you can keep your Terraform code DRY whilst not changing the way you invoke Terraform (be it via CI, Terraform Cloud, SpaceLift, env0, etc).

## Less boilerplate

TerraplateTerraplate has inheritance built in without being explicit (e.g. functions like `find_in_parent_folders()` don't need to be used).
Whilst this is very minor, it does reduce the amount of boilerplate needed in your `terraplate.hcl` configurations.
In fact, some `terraplate.hcl` files can be completely empty, because they inherit everything from parent `terraplate.hcl` files.

## Extra comments

There's a lot of things you can do with Terragrunt that you cannot do with Terraplate.
Like mentioned, we are Terragrunt fans and have been trying to find a happy place using *just* Terraform, and that's why Terraplate was created.
If you start with Terraplate and find it's not for you; that's ok, there's no lock-in as all the files are just vanilla Terraform.

If you are a Terragrunt user and find useful things missing, please raise an issue or discussion :)

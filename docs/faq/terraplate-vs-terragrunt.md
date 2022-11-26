---
title: "Terraplate vs Terragrunt"
description: "Terraplate and Terragrunt solve similar problems, in a similar way. So why create Terraplate?"
---

Terragrunt is an amazing tool, that shaped a lot of how Terraform is used today (including Terraplate).
However, there are things about Terragrunt that motivated us to write Terraplate, and those have been summarised below.
This is in no way to say that Terraplate is *better* than Terragrunt, but we do feel it is simpler.

## Native Terraform Syntax

Terragrunt does not produce native Terraform code that can be read or version controlled.
Of course in the end, Terragrunt generates native Terraform code which is invoked using the Terraform CLI, but this happens at runtime and is managed by Terragrunt.

This means:

1. You are dependent on Terragrunt to run Terraform
   - This rules out tools like [Terraform Cloud](https://cloud.hashicorp.com/products/terraform)
2. There are subtle differences in syntax between using Terraform features, like modules, which as a concept existed in Terragrunt before they existed in Terraform

## Calling none or multiple modules

Terragrunt is designed to call exactly one Terraform module from each Root Module.
Sometimes you don't want to call a Terraform module, or might want to call multiple Terraform modules from the same Root Module.

With Terraplate, you can use the templating engine to create things like the backend and providers, and then write a plain `.tf` file yourself which can do whatever you want it to.

## Keep it DRY, only

Terraplate can be used for *just* it's templating engine.
That way, you can keep your Terraform code DRY whilst not changing the way you invoke Terraform (be it via CI, Terraform Cloud, SpaceLift, env0, etc).

## Less boilerplate

Terraplate has inheritance built in without being explicit (e.g. functions like `find_in_parent_folders()` don't need to be used).
Whilst this is very minor, it does reduce the amount of boilerplate needed in your `terraplate.hcl` configurations.
In fact, some `terraplate.hcl` files can be completely empty, because they inherit everything from parent `terraplate.hcl` files.

## Extra comments

There's a lot of things you can do with Terragrunt that you cannot do with Terraplate.
Like mentioned, we are Terragrunt fans and have been trying to find a happy place using *just* Terraform, and that's why Terraplate was created.
If you start with Terraplate and find it's not for you; that's ok, there's no lock-in as all the files are just vanilla Terraform.

If you are a Terragrunt user and find useful things missing, please raise an issue or discussion :)

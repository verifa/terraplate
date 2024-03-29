---
title: "Multiple root modules"
description: ""
---

This tutorial is a light introduction to Terraplate where we take a single Terraform Root Module and split it into multiple root modules, whilst keeping things DRY.

## Example Root Module

Let's setup a basic Terraform root module where we use the `local` provider so you don't need to worry about cloud providers.

```console
# Check out the terraplate codebase containing the tutorials
git clone https://github.com/verifa/terraplate.git

# Go to the tutorial
cd terraplate/tutorials/multiple-root-modules
```

### Resources (and provider)

In there we should have a Terraform file with something like the following, which will create two files: one for dev and one for prod.
Ignore that this is stupidly simple and imagine instead you are creating VPCs, VMs, Kubernetes clusters, ... whatever you normally do!

```terraform title="main.tf"
--8<-- "tutorials/multiple-root-modules/main.tf"
```

### Backend

The `backend.tf` file defines where the Terraform state should be stored.
This config replicates the default backend which is to store the state in a local file called `terraform.tfstate`.

```terraform title="backend.tf"
--8<-- "tutorials/multiple-root-modules/backend.tf"
```

### Versions

The `versions.tf` file contains the required providers and the Terraform CLI version.

```terraform title="versions.tf"
--8<-- "tutorials/multiple-root-modules/versions.tf"
```

### Apply the configuration

Now we will apply the configuration using basic Terraform

```bash
# Initialize the root module
terraform init

# Plan the root module
terraform plan -out tfplan

# Apply based on the plan output
terraform apply tfplan

# Check output
cat prod.txt
cat dev.txt
```

Great! This should've worked. And let's imagine that it took a long time to plan, because of all your resources being inside a single Root Module and therefore a single state.

## Using Terraplate

Let's refactor this code and split the two `local_file` resources up into their own Root Modules and use Terraplate to keep things DRY.
Take a look in the `tutorials/multiple-root-modules-finished` directory for the same codebase that has been Terraplate'd.

```console title="Terraplate Structure"
# Move into the finished tutorial
cd tutorials/multiple-root-modules-finished

# Check the files we have
tree
.
├── README.md
├── local
│   ├── dev
│   │   ├── main.tf
│   │   └── terraplate.hcl
│   ├── prod
│   │   ├── main.tf
│   │   └── terraplate.hcl
│   └── terraplate.hcl
├── templates
│   └── provider_local.tmpl
└── terraplate.hcl
```

### Resource files

Let's inspect the `main.tf` files in the `local/dev` and `local/prod` environments. Note that these are identical and manually maintained (NOT currently generated by Terraplate).

```terraform title="local/dev/main.tf"
--8<-- "tutorials/multiple-root-modules-finished/local/dev/main.tf"
```

### Templates

Currently we have two templates in the `templates/` directory.

!!! info "`templates` directory is a convention"

    The `templates` directory is not required but it's a convention to keep the
    template files organized. Putting files in a `templates` does not mean
    or do anything: you still have to declare your templates using a `templates`
    block inside your Terrafiles.

They will be processed by the Go templating engine so we could set values we want based on the Root Module where it should be templated.
But for these simple files we don't need it.

```terraform title="templates/backend_local.tmpl"
--8<-- "tutorials/multiple-root-modules-finished/templates/backend_local.tmpl"
```

```terraform title="templates/provider_local.tmpl"
--8<-- "tutorials/multiple-root-modules-finished/templates/provider_local.tmpl"
```

We need to declare these templates in our Terrafiles. The backend we want to use in *every* root module so we will declare it in the root Terrafile `terraplate.hcl`.
The `local` provider we only want to use in the Terrafiles under the `local/` directory, so we place it in the `local/terraplate.hcl` Terrafile and all the child directories will inherit this template.

That takes care of the backend and providers.

### Versions

Defining the required Terraform version and `required_providers` everywhere is tiresome to do and maintain.
With Terraplate we keep the required versions at each level in the directory structure where we need them, and the child directories inherit those.

At the root level, `terraplate.hcl`, we define the Terraform CLI version.
At the `local/terraplate.hcl` directory level we declare the `local` provider.

### Terrafiles

```terraform title="terraplate.hcl"
--8<-- "tutorials/multiple-root-modules-finished/terraplate.hcl"
```

```terraform title="local/terraplate.hcl"
--8<-- "tutorials/multiple-root-modules-finished/local/terraplate.hcl"
```

```terraform title="local/dev/terraplate.hcl"
--8<-- "tutorials/multiple-root-modules-finished/local/dev/terraplate.hcl"
```

```terraform title="local/prod/terraplate.hcl"
--8<-- "tutorials/multiple-root-modules-finished/local/prod/terraplate.hcl"
```

### Apply using Terraplate

```console title="Apply with Terraplate"
# The parse command gives us a summary of the Root Modules (useful for debugging)
terraplate parse

# Let's build the templates
terraplate build

# Then we can plan (and init in the same run)
terraplate plan --init

# Finally apply the plans
terraplate plan
```

### Want to get even DRYer?

The `main.tf` file is currently the same for the `dev` and `prod` environments.
We could define a template for this, let's say under the `local/templates` directory.

```terraform title="local/templates/file.tmpl"
resource "local_file" "dev" {
  # We can use Go templates to build the value right in the file if we want!!
  content  = "env = {{ .Locals.environment }}"
  filename = "${path.module}/${local.environment}.txt"
}
```

```terraform title="local/terraplate.hcl"
template "file" {
  contents = read_template("file.tmpl")
}
```

```console
# Remove the files we are about to make DRY
rm local/dev/main.tf local/prod/main.tf

# Re-build to generate our new `file.tp.tf` files
terraplate build

# Plan and see that there should be no changes...
terraplate plan
```

## Summary

We had a *single root module* with a *single state* that we separated into *two root modules* and therefore *two separate states*.
We can now create many more root modules and the version, providers and backend are inherited and templated for us by Terraplate.
Thus, the steps for creating a new root module, such as a `staging` would be as follows:

```console title="Creating a new Root Module"
mkdir local/staging

touch local/staging/terraplate.hcl
```

And something like the following in your `terraplate.hcl` file

```terraform title="local/staging/terraplate.hcl"

locals {
  environment = "staging"
}

```

Then just add your `.tf` files, or add some more templates, and away we go!

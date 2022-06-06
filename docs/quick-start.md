---
title: "Quick Start"
description: "One page summary of how to get started with Terraplate."
---

Let's get you using Terraplate ASAP!

## Installation

See [installation instructions](./installation.md)

## Example

### Clone Terraplate

The Terraplate repository comes with some examples. Let's start with the simple one.

```console

git clone https://github.com/verifa/terraplate.git

cd terraplate/examples/simple
```

### Run Terraplate

```console
# Parse the Terrafiles and print some details
terraplate parse

# Build the templates
terraplate build

# Plan the root modules
terraplate plan

# Apply the root modules
terraplate apply
```

## Tutorials

Check the [tutorials](./tutorials/multiple-root-modules.md) for learning about how to setup a project using Terraplate

## Reference

The [complete terrafile](./reference/complete.md) reference tells you everything you can put into a Terrafile.

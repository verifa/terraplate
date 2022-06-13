---
title: "Terraplate vs Terraform Workspaces"
description: "Terraplate and Terraform workspaces help keep Terraform DRY. So how do they compare?"
lead: "Terraplate and Terraform workspaces help keep Terraform DRY. So how do they compare?"
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

## Terraform Workspaces

TODO

Terraform's own [documentation](https://www.terraform.io/language/state/workspaces#when-to-use-multiple-workspaces) also do not recommend workspaces in certain cases:

> *Workspaces alone are not a suitable tool for system decomposition, because each subsystem should have its own separate configuration and backend, and will thus have its own distinct set of workspaces.*

We know some people really enjoy workspaces, so opinions seem to vary and that's fine.

---
title: "Project Status"
description: ""
lead: ""
date: 2020-10-13T15:21:01+02:00
lastmod: 2020-10-13T15:21:01+02:00
draft: false
images: []
menu:
  docs:
    parent: "prologue"
weight: 150
toc: true
---

Terraplate is currently in **alpha**.

It is being used in production, but the functionality may change and backwards compatibility cannot be guaranteed at this time.

Terraplate does not have a cache, cleanup functionality or overwrite protection.
It's quite dumb, by design, so please be careful that you do not overwrite existing files (use Git, wink wink) and name your templated Terraform files with a suffix such as `.tp.tf` (which is the default) to add another layer of "protection".

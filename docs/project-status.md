---
title: "Project Status"
description: "Terraplate project status"
---

Terraplate is currently in **alpha**.

It is being used in production, but the functionality may change and backwards compatibility cannot be guaranteed at this time.

Terraplate does not have a cache, cleanup functionality or overwrite protection.
It's quite dumb, by design, so please be careful that you do not overwrite existing files (use Git, wink wink) and name your templated Terraform files with a suffix such as `.tp.tf` (which is the default) to add another layer of "protection".

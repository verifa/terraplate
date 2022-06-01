
locals {
  key = "value"
}

variables {
  key = "value"
}

values {
  key = "value"
}

exec {
  skip       = true
  extra_args = []

  plan {
    input    = false
    lock     = true
    out      = "tfplan"
    skip_out = false
  }
}

terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }

  required_version = ">= 1.1.0"
}

#
# NOTE: THIS FILE WAS AUTOMATICALLY GENERATED BY TERRAPLATE
#
# Terrafile: override/b/terraplate.hcl

terraform {
  required_version = ">= 1.1.0"

  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }
}

locals {
  key = "value"
}

variable "key" {
  default = "value"
}


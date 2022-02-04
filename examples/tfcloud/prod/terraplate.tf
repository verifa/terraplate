terraform {
  required_version = ">= 1.1.0"

  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }
}

variable "environment" {
  default = "prod"
}


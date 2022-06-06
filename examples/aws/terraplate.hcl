
template "backend" {
  contents = read_template("backend.tmpl")
}

template "providers" {
  contents = read_template("providers.tmpl")
}
template "common" {
  contents = read_template("common.tmpl")
}

variables {
  aws_region  = "eu-west-1"
  environment = "global"
  project     = "terraplate-aws-example"
}

exec {
  # Requires AWS auth to run, so skip running Terraform
  skip = true
}

terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.61.0"
    }
  }
}

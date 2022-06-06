
# Define a template that reads from a file
template "example" {
  contents = read_template("example.tmpl")
  # Override the target where to write the file (default: example.tp.tf)
  target    = "example_custom.tp.tf"
  condition = "{{ eq .Locals.key \"value\" }}"
}

# Define a template that embeds the content
template "embedded" {
  contents = <<-EOL
    # Content here will be templated
  EOL
}

# Define terraform locals that can be used for templating and are written to the
# terraplate.tf file
locals {
  key = "value"
  other = {
    msg = "can also be an object"
    nested = {
      list = ["and an object", "with a list"]
    }
  }
}

# Define terraform variables that can be used for templating and are written to the
# terraplate.tf file
variables {
  key = "value"
}

# Define values that can be used for templating but are *not* written to any
# terraform files
values {
  key = "value"
}

# Define an exec block configuring how terraform is executed from terraplate
exec {
  # Whether to skip running terraform. This is useful for disabling some root
  # modules but wanting to keep them in Git
  skip = false
  # Extra args arbitrary strings passed to terraform for each invocation
  extra_args = []

  # Plan controls the behaviour for running terraform plan, and subsequntly
  # affects some of the terraform apply commands
  plan {
    # Whether to enable/disable inputs (terraform's -input option)
    input = false
    # Whether to hold a state lock (terraform's -lock option)
    lock = true
    # Name of terraform plan out file that can be used as input for terraform
    # apply (terraform's -out option)
    out = "tfplan"
    # Whether to skip producing an output. Default is false (i.e. use an out file)
    skip_out = false
  }
}

# Define a terraform block for providing things like required_providers which are
# built into the terrafile.tf output file
terraform {
  # Required providers are output to the terraplate.tf file
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }

  required_version = ">= 1.1.0"
}


values {
  tfstate_file = "terraform.tfstate"
}

template "backend" {
  contents = read_template("backend.tmpl")
}

template "providers" {
  contents = read_template("providers.tmpl")
}

template "main" {
  contents = read_template("main.tmpl")
}

template "override" {
  contents = read_template("override.tmpl")
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

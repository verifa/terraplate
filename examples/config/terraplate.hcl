
template "cluster" {
  contents = read_template("cluster.tmpl")
}

template "config" {
  contents = read_template("config.tmpl")
}

values {
  tfstate_file = "terraform.tfstate"
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


template "provider_local" {
  contents = read_template("provider_local.tmpl")
}

terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "2.1.0"
    }
  }
}

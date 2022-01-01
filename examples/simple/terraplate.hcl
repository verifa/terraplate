
values {
  tfstate_file = "terraform.tfstate"
}

# Refer to the template ignore.tp.tf
template "ignore" {
  # source = "ignore.tp.tf"
  build = false
}

required_providers {
  local = {
    source  = "hashicorp/local"
    version = "2.1.0"
  }
}

required_version = ">= 1.1.0"


values {
  tfstate_file = "terraform.tfstate"
}

required_providers {
  local = {
    source  = "hashicorp/local"
    version = "2.1.0"
  }
}

required_version = ">= 1.1.0"

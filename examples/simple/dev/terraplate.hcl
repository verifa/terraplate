
variables {
  environment = "dev"
}

values {
  val = {
    "wahtever" = 123
    "another"  = 21231231
  }
}

required_providers {
  # kubernetes = {
  #   source  = "hashicorp/kubernetes"
  #   version = ">= 2.5.0"
  # }
}

# template "variables" {
#   # path = "asdasd"
#   build = true
# }

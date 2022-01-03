variables {
  aws_region  = "eu-west-1"
  environment = "global"
  project     = "terraplate-aws-example"
}

required_version = ">= 1.0.0"

required_providers {
  aws = {
    source  = "hashicorp/aws"
    version = ">= 3.61.0"
  }
}


terraform {
  backend "s3" {
    bucket  = "bucket-name"
    key     = "{{ .Path }}/terraform.tfstate"
    region  = "{{ .Variables.aws_region }}"
    encrypt = true
  }
}

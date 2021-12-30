
terraform {
  backend "local" {
    path = "{{ .Values.tfstate }}"
  }
}

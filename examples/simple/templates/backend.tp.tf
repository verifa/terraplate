
terraform {
  backend "local" {
    path = "{{ .Values.tfstate_file }}"
  }
}


template "backend" {
  contents = read_template("backend_local.tmpl")
}

terraform {
  required_version = ">= 1.0"
}

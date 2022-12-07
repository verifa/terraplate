
locals {
    config = "my-config"
}

template "config" {
    contents = read_template("config.yaml.tmpl")
    target = "config.yaml"
}


resource "local_file" "this" {
  content  = "env = ${local.environment}"
  filename = "${path.module}/${local.environment}.txt"
}

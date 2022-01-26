resource "local_file" "this" {
  content  = "env = ${var.environment}"
  filename = "${path.module}/env.txt"
}

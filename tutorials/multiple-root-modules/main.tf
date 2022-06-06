
provider "local" {
  # No configuration required
}

# Create our dev environment
resource "local_file" "dev" {
  content  = "env = dev"
  filename = "${path.module}/dev.txt"
}

# Create our prod environment
resource "local_file" "prod" {
  content  = "env = prod"
  filename = "${path.module}/prod.txt"
}

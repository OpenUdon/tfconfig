variable "name" {}

resource "example_child" "main" {
  name = var.name
}

module "grandchild" {
  source   = "./grandchild"
  for_each = toset(["one"])
}

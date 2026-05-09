variable "name" {
  default = "first"
}

variable "name" {
  default = "second"
}

resource "example_resource" "main" {
  nested {
    value = "unsupported"
  }
}

module "symbolic" {
  source = var.module_source
}

module "bare" {
  source = "."
}

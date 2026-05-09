provider "aws" {
  alias  = "west"
  region = "us-west-2"
}

module "child" {
  source = "./modules/child"

  providers = {
    aws = aws.west
  }

  name       = var.name
  count      = 1
  depends_on = [aws_security_group.root]
}

module "missing" {
  source = "./missing"
}

module "git" {
  source = "git::https://example.com/mod.git"
}

resource "aws_security_group" "root" {
  name = "root"
}

variable "name" {
  default = "app"
}

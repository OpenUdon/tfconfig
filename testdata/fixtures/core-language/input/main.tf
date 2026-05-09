resource "aws_instance" "web" {
  provider = aws.west
  ami      = data.aws_ami.base.id
  name     = local.service_name
  token    = "token-from-fixture"
  count    = length(var.names)

  depends_on = [
    data.aws_ami.base,
  ]

  lifecycle {
    prevent_destroy      = true
    create_before_destroy = false
    ignore_changes       = [tags]
    replace_triggered_by = [aws_security_group.web]

    precondition {
      condition     = var.region != ""
      error_message = "region required"
    }
  }
}

resource "aws_security_group" "web" {
  for_each = toset(var.names)
  name     = each.value
}

data "aws_ami" "base" {
  most_recent = true
}

module "registry" {
  source = "hashicorp/consul/aws"

  providers = {
    aws = aws.west
  }

  name       = local.service_name
  depends_on = [aws_security_group.web]
  for_each   = toset(var.names)
}

moved {
  from = aws_instance.old
  to   = aws_instance.web
}

import {
  to       = aws_instance.web
  id       = "i-123"
  provider = aws.west
}

removed {
  from = aws_instance.gone

  lifecycle {
    destroy = false
  }
}

check "health" {
  assert {
    condition     = data.aws_ami.base.id != ""
    error_message = "AMI missing"
  }
}

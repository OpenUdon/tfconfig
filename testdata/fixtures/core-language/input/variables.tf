variable "region" {
  type        = string
  default     = "us-east-1"
  description = "AWS region"
}

variable "api_token" {
  sensitive = true
  default   = "plain-secret"
}

variable "names" {
  type    = list(string)
  default = ["web", "worker"]
}

locals {
  service_name = "${var.region}-service"
  tags = {
    env = var.region
  }
}

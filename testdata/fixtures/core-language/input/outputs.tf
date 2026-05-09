output "service_name" {
  value      = local.service_name
  depends_on = [aws_instance.web]
}

output "token" {
  sensitive = true
  value     = "inline-password"
}

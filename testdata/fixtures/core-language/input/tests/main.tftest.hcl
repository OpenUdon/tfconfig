run "plan_core" {
  command = plan

  variables {
    region = "us-west-2"
  }

  assert {
    condition     = output.service_name != ""
    error_message = "missing service name"
  }
}

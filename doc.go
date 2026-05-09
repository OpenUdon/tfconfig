// Package tfconfig loads Terraform/OpenTofu configuration into a static,
// review-oriented fact model.
//
// The package is intentionally static analysis only. It must not execute
// provider plugins, initialize backends, load state, refresh, plan, or apply.
package tfconfig

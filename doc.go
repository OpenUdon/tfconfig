// Package tfconfig loads Terraform/OpenTofu configuration into a static,
// review-oriented fact model.
//
// The initial public contract is identified by StaticV1. OpenUdon consumes the
// Go model as the primary API; Document.JSON and Document.JSONIndent provide the
// deterministic public JSON projection for export, fixtures, and review.
//
// LoadDir is the public local-static loader entrypoint. It loads .tf, .tofu,
// .tf.json, .tofu.json, override, and test files from a root module directory
// and recursively loads direct readable local child modules without downloading
// modules or executing Terraform/OpenTofu.
//
// The package is intentionally static analysis only. It must not execute
// provider plugins, initialize backends, load state, refresh, plan, or apply.
package tfconfig

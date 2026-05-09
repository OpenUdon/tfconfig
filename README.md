# tfconfig

`tfconfig` is the OpenUdon-owned static Terraform/OpenTofu configuration parser
package and tool.

Module path:

```text
github.com/OpenUdon/tfconfig
```

The target contract is a deterministic `tfconfig.static.v1` fact model that
OpenUdon can consume when scaffolding reviewable UWS packages from
Terraform/OpenTofu configuration and OpenAPI documents.

This package is static analysis only. It does not run provider plugins,
initialize backends, load state, refresh, plan, or apply.

OpenTofu-derived files are MPL-2.0-covered and must preserve their upstream
headers. See [AGENTS.md](AGENTS.md) and [UPSTREAM.md](UPSTREAM.md).

## OpenTofu Sync

Allowlisted upstream files are listed in [sync/opentofu-files.tsv](sync/opentofu-files.tsv).
To sync from a local OpenTofu checkout:

```bash
OPENTOFU_DIR=../opentofu ./scripts/sync-opentofu.sh
go test ./...
go vet ./...
```

GitHub Actions also includes a weekly/manual workflow that runs the sync and
opens a review-required pull request when files change.

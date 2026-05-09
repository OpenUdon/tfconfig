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
headers. See [AGENTS.md](AGENTS.md), [UPSTREAM.md](UPSTREAM.md), and
[THIRD_PARTY.md](THIRD_PARTY.md).

## OpenTofu Sync

Allowlisted upstream files are listed in [sync/opentofu-files.tsv](sync/opentofu-files.tsv).
Raw upstream snapshots should normally land under `_upstream/opentofu/...`, which
the Go tool ignores. Compile-ready parser code should then be adapted into
normal `tfconfig` packages after review.

To sync from a local OpenTofu checkout:

```bash
OPENTOFU_DIR=../opentofu ./scripts/sync-opentofu.sh
go test ./...
go vet ./...
```

GitHub Actions also includes a weekly/manual workflow that runs the sync and
opens a review-required pull request when files change.

The workflow is scheduled to run every Monday at 06:17 UTC:

```yaml
schedule:
  - cron: "17 6 * * 1"
```

You can also run it manually from GitHub:

1. Open `OpenUdon/tfconfig` on GitHub.
2. Go to **Actions**.
3. Select **Sync OpenTofu Static Config Sources**.
4. Click **Run workflow**.
5. Leave `opentofu_ref` as `main`, or enter a specific OpenTofu tag or commit.

If allowlisted OpenTofu files changed, the workflow opens or updates a
review-required sync pull request. If no pull request appears, the mirror is
already current for the selected OpenTofu ref.

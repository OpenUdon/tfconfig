# Tech Stack

## Language And Module

- Primary language: Go.
- Module path: `github.com/OpenUdon/tfconfig`.
- Public API: `LoadDir`, `LoadDirWithOptions`, and `tfconfig.static.v1` model
  structs, including optional collection-shape facts.
- CLI: `cmd/tfconfig` for deterministic JSON export/debug output.

## Key Artifacts

- `_upstream/opentofu/`: ignored raw OpenTofu source snapshots.
- `sync/opentofu-files.tsv`: allowlist of upstream files to mirror.
- `scripts/sync-opentofu.sh`: local sync helper.
- `.github/workflows/sync-opentofu.yml`: weekly/manual upstream sync using the
  Node 24 action majors `actions/checkout@v7`, `actions/setup-go@v7`, and
  `peter-evans/create-pull-request@v8`, an explicit nested `go.sum` cache path,
  and the `TFCONFIG_SYNC_TOKEN` repository secret for review-required pull
  requests. The default `GITHUB_TOKEN` remains read-only.
- `UPSTREAM.md`, `THIRD_PARTY.md`, `licenses/opentofu-MPL-2.0.txt`:
  provenance and license records.
- `docs/static-v1.md`: public model contract.
- `memory-bank/`: active parser project memory and milestone status.
- `../skills/harness/tackle-memory-bank-api-loop`: optional unattended runner
  for canonical `status-<LANE><NN>.md` task ledgers.

## Commands

```bash
go test ./...
go vet ./...
git diff --check
go run ./cmd/tfconfig --config-dir ./tf
OPENTOFU_DIR=../opentofu ./scripts/sync-opentofu.sh
actionlint .github/workflows/sync-opentofu.yml
../skills/harness/tackle-memory-bank-api-loop --model lane-audit .
```

When verifying downstream compatibility after parser changes:

```bash
(cd ../openudon && go test ./internal/tfconvert)
```

## Dependency Policy

- Keep local sibling development in the parent `go.work`, not committed
  `replace ../...` directives.
- Keep raw OpenTofu snapshots outside compile-ready package directories unless
  intentionally adapted.
- Preserve MPL-2.0 headers and provenance for copied or adapted OpenTofu files.
- Prefer deterministic parser fixtures and corpus tests over live Terraform or
  provider behavior.
- Keep the static function context explicitly whitelisted. P01 permits only
  literal `toset`; no filesystem, environment, time/random, provider, network,
  variable/local resolution, or runtime function enters the evaluator.

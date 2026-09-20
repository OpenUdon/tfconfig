# Status M02 — tfconfig Upstream Mirror Bootstrap And Provenance

State of M02 items. See [milestone.md](milestone.md) for milestone scope and
acceptance criteria.

Status markers:

| Symbol | Suggested Status | Interpretation |
|---|---|---|
| `[ ]` | Pending | Item not started or pending action. |
| `[+]` | Completed | Item finished or done. |
| `[~]` | In Progress | Item is being worked on. |
| `[!]` | Blocked | Item requires attention or is on hold. |
| `[X]` | Cancelled | Item is no longer needed. |

Each table row is a commit unit once implementation begins: after flipping a
row to `[+]`, verify the change, update the memory bank/docs, and make a scoped
commit before starting the next row. If multiple rows are inseparable, use one
coherent commit and name every covered row in the handoff.

| Item | State | Notes |
|---|---|---|
| Module identity established | `[+]` | `../tfconfig` uses module path `github.com/OpenUdon/tfconfig` |
| Sync allowlist established | `[+]` | `sync/opentofu-files.tsv` lists intentionally mirrored OpenTofu files |
| Raw upstream snapshots synced | `[+]` | Selected files are under `_upstream/opentofu/...` |
| OpenTofu runtime boundary preserved | `[+]` | `internal/tofu` remains excluded |
| Sync script installed | `[+]` | `scripts/sync-opentofu.sh` reproduces the mirror from a local OpenTofu checkout |
| GitHub sync workflow installed | `[+]` | Weekly/manual workflow uses Node 24 action majors, keeps `GITHUB_TOKEN` read-only, and opens or updates review PRs for allowlisted changes through the dedicated `TFCONFIG_SYNC_TOKEN` |
| Provenance docs present | `[+]` | `UPSTREAM.md` records upstream commit, source paths, destinations, and notes |
| Third-party notices present | `[+]` | `THIRD_PARTY.md` and MPL license text document OpenTofu licensing |
| Operator sync docs present | `[+]` | README explains manual and scheduled sync behavior |
| Bootstrap gates verified | `[+]` | `go test ./...`, `go vet ./...`, and `git diff --check` pass in `../tfconfig` |

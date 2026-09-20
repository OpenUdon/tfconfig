# Status M07 — OpenUdon Convert TF Adapter Design

State of M07 items. See [milestone.md](milestone.md) for milestone scope and
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
| OpenUdon command integration point identified | `[+]` | `docs/terraform-openapi-conversion.md` defines `cmd/openudon/main.go` routing to reusable `internal/tfconvert` logic for M08 |
| CLI flag contract finalized | `[+]` | Contract defines defaults and validation for `--config-dir`, repeatable `--openapi`, resource-required `--action`, repeatable `--target`, `--out`, and `--strict` |
| OpenAPI input behavior finalized | `[+]` | Contract defines required unique `id=PATH`, per-ID local loading/indexing, ID sorting, namespaced operation refs, and diagnostics for missing/unreadable/malformed inputs |
| Output directory and draft artifact layout finalized | `[+]` | Contract defines default `./.openudon/convert` layout with `project.md`, `workflows/intent.hcl`, diagnostics JSON/Markdown, and review notes |
| Managed resource action policy finalized | `[+]` | Contract requires explicit global `create/update/delete/replace` action for selected managed resources and forbids inferred side effects |
| Target filtering semantics finalized | `[+]` | Contract defines exact matching on adapter-computed full addresses from `tfconfig` module/object addresses plus deterministic unmatched-target diagnostics |
| Data-source mapping rules finalized | `[+]` | Contract limits data sources to read/list candidates, never requires `--action`, and emits unresolved operation TODOs when needed |
| Resource scaffold mapping rules finalized | `[+]` | Contract maps explicit resource actions to draft review scaffolds, with `replace` as delete/create slots and deterministic TODO operation IDs when no confident operation exists |
| Provider alias binding rules finalized | `[+]` | Contract normalizes provider config addresses such as `provider.aws` and `provider.aws.west` into distinct symbolic binding names |
| Symbolic value preservation rules finalized | `[+]` | Contract preserves variables, locals, outputs, references, expressions, `count`, `for_each`, dependencies, lifecycle, and module facts symbolically |
| OpenAPI operation matching policy finalized | `[+]` | Contract uses narrowed `apitools` inventory/index/summary/selection APIs with namespaced `{openapi_id, operation_id}` refs and deterministic no-match/ambiguous TODO behavior |
| Diagnostics and strict-mode policy finalized | `[+]` | Contract defines stable diagnostic fields, JSON/Markdown outputs, TODO IDs, sort order, and strict failure conditions |
| Redaction and sensitive variable design finalized | `[+]` | Contract converts `tfconfig` sensitive values and sensitive candidates into symbolic sensitive OpenUdon inputs/review variables without emitting secret-like literals |
| Deterministic output rules finalized | `[+]` | Contract sorts OpenAPI docs, targets, bindings, namespaced operation candidates, diagnostics, and TODOs while avoiding host/runtime-derived values |
| Boundary check design finalized | `[+]` | Contract keeps OpenUdon off OpenTofu internals and requires only narrowed `apitools` APIs guarded by `openudon check-apitools-boundary` |

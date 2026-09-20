# Status M08 — OpenUdon Convert TF Initial Implementation

State of M08 items. See [milestone.md](milestone.md) for milestone scope and
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
| `openudon convert tf` command wired | `[+]` | `../openudon/cmd/openudon` now routes `openudon convert tf` to a thin CLI parser with help text and repeatable flags |
| `tfconfig` Go API dependency added | `[+]` | `../openudon` consumes `github.com/OpenUdon/tfconfig` through the Go API with no OpenTofu internals; local sibling overrides belong in the parent workspace, not committed `go.mod` |
| Config directory loading implemented | `[+]` | Adapter loads `--config-dir` through `tfconfig.LoadDir` and converts document/module parser diagnostics into conversion diagnostics |
| Repeatable OpenAPI loading implemented | `[+]` | Adapter parses repeatable `--openapi id=PATH`, loads each ID through `apitools.BuildOperationInventory`, and indexes per ID with `NewOperationIndex` |
| Managed resource action validation implemented | `[+]` | Adapter validates `create/update/delete/replace` and emits strict-failure diagnostics when selected managed resources lack a valid action |
| Target filtering implemented | `[+]` | Adapter filters by repeatable computed Terraform addresses and emits deterministic unmatched-target diagnostics |
| Data-source draft mapping implemented | `[+]` | Adapter maps data sources to read/list operation slots with deterministic TODOs for unresolved or ambiguous matches |
| Managed resource draft mapping implemented | `[+]` | Adapter maps explicit resource actions to draft steps, including delete/create slots for `replace` and TODOs for unresolved operations |
| Provider alias symbolic bindings emitted | `[+]` | Adapter emits normalized provider binding names and review/security context without resolving credentials |
| Variable/local/output symbolic facts emitted | `[+]` | Adapter preserves variables as intent inputs and locals/outputs as symbolic review facts without inventing runtime values |
| Draft `project.md` emitted | `[+]` | Adapter writes draft project summary with inputs, selected objects, diagnostics, and unapproved review posture |
| Draft `workflows/intent.hcl` emitted | `[+]` | Adapter writes initial OpenUdon intent scaffolding with symbolic inputs, security bindings, and draft operation/TODO steps |
| Diagnostics JSON and Markdown emitted | `[+]` | Adapter writes deterministic `expected/diagnostics.json` and `expected/diagnostics.md` with strict-failure metadata |
| Review notes emitted | `[+]` | Adapter writes `expected/review.md` summarizing operation mappings, provider bindings, symbolic facts, and redaction TODOs |
| Strict mode failure implemented | `[+]` | `--strict` writes artifacts and exits non-zero when strict-failure diagnostics remain |
| Initial adapter tests added | `[+]` | Tests cover successful artifact generation, ambiguous TODOs, redaction, strict failures, duplicate operation IDs across OpenAPI IDs, CLI wiring, and deterministic output |

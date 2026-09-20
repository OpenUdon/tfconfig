# Status M10 — Parser Stewardship

State of `tfconfig` stewardship items. See [milestone.md](milestone.md) for
milestone scope and acceptance criteria.

Status markers:

| Symbol | Suggested Status | Interpretation |
|---|---|---|
| `[ ]` | Pending | Item not started or pending action. |
| `[+]` | Completed | Item finished or done. |
| `[~]` | In Progress | Item is being worked on. |
| `[!]` | Blocked | Item requires attention or is on hold. |
| `[X]` | Cancelled | Item is no longer needed. |

| Item | State | Notes |
|---|---|---|
| Sync stewardship process documented | `[+]` | [release-stewardship.md](../docs/release-stewardship.md) records the parser sync and review path |
| Provenance checklist maintained | `[+]` | Checklist covers `opentofu-files.tsv`, sync script, `UPSTREAM.md`, `THIRD_PARTY.md`, MPL license copy, and compile-ready adaptation boundaries |
| OpenTofu parser change review notes maintained | `[+]` | Parser review focuses on static facts, diagnostics, source ranges, symbols, sensitivity, modules, and deterministic ordering |
| Unsupported behavior docs maintained | `[+]` | Provider execution, module downloads, backends, state, plan/apply, test execution, credentials, and OpenUdon package behavior remain out of scope |
| Static-only limitation diagnostics reviewed | `[+]` | Missing modules, unsupported sources, parser diagnostics, and schema-less provider facts remain visible in the static model |
| Deterministic output regression tests maintained | `[+]` | Model, loader, and fixture tests cover deterministic JSON and canonical ordering |
| Redaction regression tests maintained | `[+]` | Sensitive value model/load tests and fixture leak checks protect safe public JSON |
| Local/missing module regression tests maintained | `[+]` | Local module tree, unavailable source, unreadable source, cycle guard, and module manifest tests cover module boundaries |
| Optional OpenTofu corpus checks maintained | `[+]` | Equivalence and valid-modules corpus tests run when sibling `../opentofu` fixtures are available |
| Downstream converter smoke maintained | `[+]` | OpenUdon `internal/tfconvert` smoke remains the parser compatibility check after static model changes |

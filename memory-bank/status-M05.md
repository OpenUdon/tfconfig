# Status M05 — tfconfig Local Module Tree Loading

State of M05 items. See [milestone.md](milestone.md) for milestone scope and
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
| Local module source classifier implemented | `[+]` | `../tfconfig` classifies direct local, missing, remote/downloader-backed, and unsupported symbolic sources |
| Local path normalization implemented | `[+]` | Local module sources resolve from the calling module directory and record normalized module dirs |
| Direct local module source detection implemented | `[+]` | Only direct filesystem sources (`./`, `../`, Windows-style equivalents, and absolute paths) are loaded |
| Recursive local module walker implemented | `[+]` | Nested local modules load recursively with full stable module addresses |
| Module input/provider mapping preservation implemented | `[+]` | Existing symbolic inputs, `count`, `for_each`, `depends_on`, and provider mappings remain on module calls |
| Missing/remote module diagnostics implemented | `[+]` | Missing local paths and registry/Git/HTTP/S3/OCI/downloader-backed sources produce placeholder modules with diagnostics; no downloads or `tofu init` |
| Recursion and cycle guard implemented | `[+]` | Active module path stack prevents recursive cycles and emits deterministic diagnostics |
| Module manifest reading remains deferred | `[+]` | `.terraform/modules/modules.json` is intentionally ignored and covered by regression test |

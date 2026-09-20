# Status M06 — Parser Fixture Corpus

State of M06 items. See [milestone.md](milestone.md) for milestone scope and
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
| Golden fixture harness added | `[+]` | `../tfconfig` fixture tests compare deterministic `tfconfig.static.v1` output and diagnostics under `testdata/fixtures` |
| Multi-file and format fixtures added | `[+]` | `.tf`, `.tofu`, JSON variants, overrides, and deterministic source ranges are covered |
| Override ordering fixtures added | `[+]` | Primary/override merge behavior and missing-base override diagnostics are locked with expected output |
| Named value fixtures added | `[+]` | Variables, defaults, locals, outputs, references, and sensitivity metadata are covered |
| Provider fixtures added | `[+]` | Required providers, aliases, provider references, and module `providers = { ... }` mappings are covered |
| Resource/data-source fixtures added | `[+]` | Lifecycle, `depends_on`, `count`, `for_each`, moved/import/removed/check/test facts are covered |
| `count` and `for_each` expression fixtures added | `[+]` | Symbolic meta-argument expressions are locked in fixture output |
| Module-call fixtures added | `[+]` | Module source, inputs, dependencies, and provider mappings are covered through local and remote module calls |
| Module fixtures added | `[+]` | Local modules, nested local modules, missing modules, and remote-source diagnostics are covered |
| Diagnostic fixtures added | `[+]` | Duplicate declarations, invalid syntax, unsupported blocks, source diagnostics, and module diagnostics are covered |
| Source range fixtures added | `[+]` | Normalized paths and source ranges are locked across HCL and JSON inputs |
| Safe value projection fixtures added | `[+]` | Known likely-secret literals are redacted in public JSON and checked by regression test |
| Regression tests wired | `[+]` | `go test ./...` runs the fixture corpus and fails on output drift |

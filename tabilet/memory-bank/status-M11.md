# Status M11 — OpenTofu Equivalence Fixture Hardening

State of M11 items. See [milestone.md](milestone.md) for milestone scope and
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
| OpenTofu equivalence corpus wired | `[+]` | `fixture_test.go:TestOpenTofuEquivalenceFixtureCorpus` loads `../opentofu/testing/equivalence-tests/tests` when present and skips when absent |
| Corpus size guard added | `[+]` | Test expects 49 fixture directories so upstream fixture drift is reviewed instead of silently ignored |
| Static-only fixture policy preserved | `[+]` | Corpus test uses `tfconfig.LoadDir` only; it does not run OpenTofu, install providers, read state, plan, apply, or execute `tfcoremock` |
| Schema-less nested provider blocks preserved | `[+]` | `decode.go` flattens unknown provider/resource/data nested blocks into deterministic dotted config paths |
| Repeated nested block pathing implemented | `[+]` | Repeated nested blocks use source-order indexes such as `first_block[0].id` and `first_block[1].id` |
| Built-in meta block separation maintained | `[+]` | Resource `lifecycle` remains modeled in explicit lifecycle fields rather than generic config attributes |
| Focused parser fixtures updated | `[+]` | Existing parser tests now assert nested provider block preservation and refreshed expected JSON |
| Public contract docs updated | `[+]` | `docs/static-v1.md` documents schema-less nested provider block preservation |
| Regression checks verified | `[+]` | `tfconfig` passes `go test ./...`, `go vet ./...`, `git diff --check`, and a 49-fixture smoke pass with zero failures |
| Downstream converter smoke verified | `[+]` | `../openudon` passes `go test ./internal/tfconvert` against the workspace `tfconfig` checkout |

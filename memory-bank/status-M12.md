# Status M12 — OpenTofu Valid Modules Corpus Hardening

State of M12 items. See [milestone.md](milestone.md) for milestone scope and
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
| OpenTofu valid-modules corpus wired | `[+]` | `fixture_test.go:TestOpenTofuValidModulesFixtureCorpus` loads `../opentofu/internal/configs/testdata/valid-modules` when present and skips when absent |
| Corpus size guard added | `[+]` | Test expects 34 fixture directories so upstream fixture drift is reviewed instead of silently ignored |
| Static-only fixture policy preserved | `[+]` | Corpus test uses `tfconfig.LoadDir` only; it does not run OpenTofu, initialize backends, install providers, read state, plan, apply, or execute tests |
| Initial clean fixtures identified | `[+]` | The pre-M12 corpus had 18 zero-error fixture directories that now remain covered by the corpus regression test |
| Initial gap fixtures categorized | `[+]` | The 16 initial failures grouped into ephemeral blocks, backend/cloud/provider_meta blocks, required-provider overrides, richer test-file decoding, and the intentional missing local module fixture |
| Ephemeral resources decoded | `[+]` | `tfconfig.static.v1` now exposes `ephemeral_resources` with static config, provider, dependency, reference, and range facts |
| Backend/cloud/provider_meta decoded | `[+]` | Terraform `backend`, `cloud`, and `provider_meta` blocks are preserved as static config facts without backend or cloud initialization |
| Required provider overrides handled | `[+]` | Override files can replace existing required-provider declarations without false duplicate diagnostics |
| Richer test-file blocks decoded | `[+]` | Test files now preserve top-level `variables` and run-local `module`, `variables`, `plan_options`, and `assert` blocks for HCL and JSON fixtures |
| Intentional override-module diagnostic retained | `[+]` | `override-module` is expected to report only `module_source_missing` because `tfconfig` recursively checks local module sources |
| Public contract docs updated | `[+]` | `docs/static-v1.md` documents the expanded static facts and non-execution boundary |
| Regression checks verified | `[+]` | `tfconfig` passes `go test ./...`, `go vet ./...`, `git diff --check`, and a 34-fixture valid-modules smoke with 33 zero-error fixtures plus the expected `override-module` diagnostic; downstream OpenUdon passed `go test ./...`, `go vet ./...`, `make check`, `git diff --check`, and the same OpenUdon gates with `GOWORK=off` |
| Deep review findings addressed | `[+]` | Follow-up fixes corrected backend/cloud override semantics, restored malformed HCL test block diagnostics, and added semantic corpus assertions for backend/cloud, provider_meta, ephemeral, and test-file facts |

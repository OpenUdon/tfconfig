# Status M04 — tfconfig Local-Static Loader

State of M04 items. See [milestone.md](milestone.md) for milestone scope and
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
| Public `LoadDir` entrypoint implemented | `[+]` | `../tfconfig.LoadDir` and `LoadDirWithOptions` load a root config directory and return canonical `tfconfig.static.v1` |
| Multi-file directory discovery implemented | `[+]` | `../tfconfig/load.go` discovers `.tf`, `.tofu`, `.tf.json`, `.tofu.json`, primary, override, and root/test-directory test files |
| OpenTofu-derived file ordering preserved | `[+]` | Discovery follows OpenTofu primary-before-override ordering and ignores `.tf` files with parallel `.tofu` alternatives |
| Parser diagnostics and source ranges implemented | `[+]` | HCL parse/decode diagnostics are projected into model diagnostics with normalized source paths and ranges |
| Named values decoded | `[+]` | Variables, locals, and outputs are decoded with defaults, expressions, references, sensitivity markers, and ranges |
| Terraform block and required providers decoded | `[+]` | `terraform.required_version` and `required_providers` facts are decoded into the module model |
| Provider configs decoded | `[+]` | Provider configs, aliases, config attributes, provider refs, and references are decoded |
| Resources and data sources decoded | `[+]` | Managed resources and data sources are decoded with addresses, types, names, config attributes, and ranges |
| Lifecycle/dependency/meta-arguments decoded | `[+]` | Lifecycle, `depends_on`, `count`, and `for_each` are preserved symbolically with references |
| Module-call facts decoded | `[+]` | Module calls, sources, inputs, provider mappings, and meta-arguments are decoded without loading child module directories |
| Structural blocks decoded | `[+]` | Moved, import, removed, check, and test run facts are decoded |
| Expression references extracted | `[+]` | Attribute and meta-argument expressions preserve symbolic text and HCL traversal references |
| Deterministic Go model output returned | `[+]` | `LoadDir` returns `Document.Canonical()` and loader tests cover deterministic JSON |
| CLI JSON output implemented | `[+]` | `../tfconfig/cmd/tfconfig` prints deterministic `tfconfig.static.v1` JSON for export/debug use |

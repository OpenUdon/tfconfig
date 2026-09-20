# Status M03 — tfconfig.static.v1 Model Contract

State of M03 items. See [milestone.md](milestone.md) for milestone scope and
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
| Primary Go model specified | `[+]` | `../tfconfig/model.go` defines `tfconfig.static.v1` structs and versioning as the main OpenUdon consumption boundary |
| CLI JSON projection specified | `[+]` | `Document.JSON` and `Document.JSONIndent` define stable export/debug JSON from the Go model |
| Version and root metadata specified | `[+]` | `Document`, `SourceRoot`, `SourceFile`, and `NewDocument` cover model version, root directory, normalized source roots, and producer metadata |
| Source file and source range model specified | `[+]` | `SourceFile`, `SourceRange`, and `Position` cover normalized paths, ranges, and stable source identifiers |
| Module tree model specified | `[+]` | `Module` and `ModuleStatus` cover module address, source, directory, parent/child relationships, and load status |
| Variable, local, and output model specified | `[+]` | `Variable`, `Local`, and `Output` cover defaults, descriptions, sensitivity markers, expressions, references, and source ranges |
| Provider requirement/config/alias/reference model specified | `[+]` | `ProviderRequirement`, `ProviderConfig`, `ProviderRef`, and `ProviderMapping` cover providers, aliases, refs, and static metadata |
| Resource and data-source model specified | `[+]` | `Resource`, `DataSource`, and `Attribute` cover Terraform addresses, provider refs, attributes, lifecycle, references, and source ranges |
| Module-call and provider-mapping model specified | `[+]` | `ModuleCall` and `ProviderMapping` cover source, inputs, provider mappings, `count`, `for_each`, and dependencies |
| Lifecycle/dependency/count/for_each facts specified | `[+]` | `Lifecycle`, `Reference`, and meta-argument fields cover symbolic lifecycle, `depends_on`, `count`, and `for_each` facts |
| Moved/import/removed/check/test fact model specified | `[+]` | `MovedBlock`, `ImportBlock`, `RemovedBlock`, `CheckBlock`, `CheckRule`, `TestFile`, and `TestRun` cover structural facts |
| Deterministic ordering rules specified | `[+]` | `Document.Canonical`, module normalization, and JSON tests sort modules, files, declarations, resources, references, and diagnostics |
| Expression/reference representation specified | `[+]` | `Value`, `ValueKind`, and `Reference` distinguish literals, symbolic expressions, references, unknowns, collections, and redacted values |
| Safe value projection specified | `[+]` | `Value.MarshalJSON` redacts likely-secret literals marked with `SensitiveCandidate` from public JSON |
| Contract examples or golden JSON added | `[+]` | `docs/static-v1.md` and `model_test.go` include contract examples and deterministic JSON projection coverage |

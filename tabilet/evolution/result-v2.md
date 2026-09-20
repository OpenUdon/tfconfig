# Result — v2

State captured by [`prompt-v2`](prompt-v2.md).

## Shape Today

- `../tfconfig` now uses the local OpenTofu equivalence fixture corpus as an
  optional parser regression suite when `../opentofu` is available.
- The corpus is used only through `tfconfig.LoadDir`; no OpenTofu commands,
  provider installs, state reads, plans, applies, or `tfcoremock` executions are
  part of the test.
- All 49 fixture directories currently parse without `tfconfig` error
  diagnostics in the local workspace.

## Parser Position

- Provider, managed resource, and data source bodies remain schema-less at the
  `tfconfig` boundary.
- Unknown nested provider blocks are preserved as deterministic dotted config
  paths, with source-order indexes for repeated blocks.
- Built-in meta blocks such as resource `lifecycle` continue to use explicit
  model fields.

## Cross-Repo Position

- `../tfconfig` owns the static parser behavior and the equivalence corpus
  regression test.
- `../openudon` continues to consume only the public `tfconfig` API.
- `tfconfig` records M11 scope, status, and operating guidance for future
  parser stewardship in its local memory bank.

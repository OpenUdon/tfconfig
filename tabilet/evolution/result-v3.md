# Result — v3

State captured by [`prompt-v3`](prompt-v3.md).

## Shape Today

- `../tfconfig` now uses the local OpenTofu `valid-modules` fixture corpus as an
  optional parser regression suite when `../opentofu` is available.
- The corpus is used only through `tfconfig.LoadDir`; no OpenTofu commands,
  backend initialization, provider installs, state reads, plans, applies, or
  test executions are part of the test.
- All 34 fixture directories are currently covered in the local workspace. The
  static parser expects zero error diagnostics for 33 directories and only the
  intentional `module_source_missing` diagnostic for `override-module`.

## Parser Position

- Terraform `backend`, `cloud`, and `provider_meta` blocks are preserved as
  static review facts without runtime initialization.
- `ephemeral` blocks are preserved as static resource-like declarations without
  provider execution or lifetime claims.
- Test files preserve top-level `variables` and run-local `module`,
  `variables`, `plan_options`, and `assert` blocks.
- Required-provider overrides are handled as override replacements rather than
  false duplicate declarations.

## Cross-Repo Position

- `../tfconfig` owns the static parser behavior and both OpenTofu corpus
  regression tests.
- `../openudon` continues to consume only the public `tfconfig` API.
- `tfconfig` records M12 scope, status, and operating guidance for future
  parser stewardship in its local memory bank.

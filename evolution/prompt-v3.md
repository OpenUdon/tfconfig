# Prompt — v3

The directional prompt for the current roadmap slice after M12.

## Direction

Keep the Terraform/OpenTofu conversion path static and review-first, while
hardening `tfconfig.static.v1` against OpenTofu's `valid-modules` config loader
fixture corpus.

## Why This Direction

- `valid-modules` covers static language and loader constructs that the
  equivalence corpus did not stress, including `ephemeral`, backend/cloud
  declarations, `provider_meta`, provider requirement overrides, and richer test
  file shapes.
- These constructs are review-relevant source facts even though `tfconfig` must
  not execute providers, initialize backends, contact Terraform Cloud, run
  tests, or load state.
- Preserving the facts explicitly gives OpenUdon reviewers a better handoff
  surface without weakening the static boundary.

## Constraints

- Use `valid-modules` fixtures as static config directories only.
- Do not run OpenTofu, initialize backends, install providers, read state, plan,
  apply, or execute tests.
- Keep missing local module sources diagnostic-only.
- Preserve backend, cloud, provider meta, ephemeral, and test-file facts
  symbolically with source ranges and deterministic JSON.
- Keep OpenUdon on the public `tfconfig` API and off OpenTofu internals.

## Out Of Scope

- Backend initialization or cloud API behavior.
- Provider schema interpretation.
- Terraform/OpenTofu runtime validity, planning, apply, or test execution
  claims.

# Prompt — v2

The directional prompt for the current roadmap slice after M11.

## Direction

Keep the Terraform/OpenTofu conversion path static and review-first, while
hardening `tfconfig.static.v1` against realistic provider-shaped HCL from the
OpenTofu equivalence fixture corpus.

## Why This Direction

- Provider schemas remain out of scope, but real provider configuration often
  uses nested blocks rather than only attributes.
- Rejecting schema-less nested provider blocks loses static facts that OpenUdon
  reviewers need to inspect.
- OpenTofu's equivalence fixtures are useful parser corpus inputs even though
  their original purpose is runtime plan/apply equivalence.

## Constraints

- Use equivalence fixtures as static config directories only.
- Do not run OpenTofu, install providers, read state, plan, apply, or execute
  `tfcoremock`.
- Preserve unknown nested provider/resource/data blocks deterministically as
  static config facts.
- Keep built-in Terraform/OpenTofu meta blocks in explicit model fields.
- Keep OpenUdon on the public `tfconfig` API and off OpenTofu internals.

## Out Of Scope

- Provider schema interpretation.
- Terraform/OpenTofu runtime equivalence claims.
- State, refresh, plan, apply, or provider plugin behavior.

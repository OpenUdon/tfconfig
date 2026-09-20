# Product

`tfconfig` is the static Terraform/OpenTofu configuration parser package and
CLI used by OpenUdon conversion. It turns local configuration directories into a
deterministic `tfconfig.static.v1` fact model without executing Terraform,
OpenTofu, providers, backends, state, plans, or credentials.

## Users

- OpenUdon maintainers consuming static Terraform/OpenTofu facts through the Go
  API.
- Parser maintainers syncing and adapting selected OpenTofu configuration
  loader behavior.
- Reviewers comparing deterministic JSON fixtures, diagnostics, source ranges,
  and likely-secret redaction.
- Agents hardening static parser behavior against local OpenTofu fixture
  corpora.

## Core Concepts

- **Config directory**: one Terraform/OpenTofu module loaded from `.tf`,
  `.tofu`, `.tf.json`, `.tofu.json`, override, and test files.
- **Static fact model**: the public `tfconfig.static.v1` Go model plus stable
  JSON projection.
- **Local module tree**: recursively loaded direct local module sources only;
  remote, missing, registry, Git, HTTP, S3, OCI, and symbolic sources remain
  diagnostics.
- **Symbolic value**: expressions and references preserved without pretending to
  know runtime values.
- **Bounded instance fact**: numeric count and wholly known object/map/set
  collections retain typed shape; only literal `toset` is evaluated, while
  runtime-dependent expressions remain symbolic.
- **Safe projection**: likely-secret literals are marked as sensitive
  candidates and redacted from public JSON.

## Scope

- Static directory and local-module loading.
- OpenTofu-derived parser source mirror, provenance, and MPL-2.0 license
  boundary.
- Public parser/model/CLI contract for `github.com/OpenUdon/tfconfig`.
- Deterministic fixtures and optional local OpenTofu corpus coverage.
- Parser diagnostics, source ranges, static declarations, symbolic expressions,
  references, and safe public JSON.
- Typed collection shape and a minimal pure-literal whitelist for downstream
  review adapters that must reject unsupported instance semantics precisely.

## Non-Goals

- Provider plugin execution or provider schema loading.
- `tofu init`, module downloads, backend initialization, state, refresh, plan,
  apply, import, or test execution.
- OpenAPI operation mapping or UWS package generation; those belong in
  `../openudon` and `../apitools`.
- Credential resolution, secret storage, or production side effects.

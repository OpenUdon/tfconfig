# AGENTS.md

## Purpose

`tfconfig` is the OpenUdon-owned static Terraform/OpenTofu configuration parser
package/tool. Its job is to turn local Terraform/OpenTofu configuration into a
stable, review-oriented fact model that OpenUdon can consume when scaffolding
UWS packages from `openudon convert tf`.

Module path:

```text
github.com/OpenUdon/tfconfig
```

This repository exists because real Terraform/OpenTofu IaC is directory-based,
not single-file HCL. A useful parser must understand multi-file loading,
override ordering, `.tf`/`.tofu` and JSON variants, modules, providers,
variables, resources, data sources, lifecycle metadata, moved/import/removed
blocks, checks, tests, diagnostics, and static module trees.

## Boundary

- `../opentofu` is the upstream reference implementation and source for
  MPL-2.0-derived static configuration loading code.
- `../tfconfig` owns a focused public parser/model/CLI boundary for OpenUdon.
- `../tofu` is a design and spike workspace for the broader
  Terraform/OpenAPI-to-OpenUdon conversion work.
- `../openudon` consumes `tfconfig` output plus OpenAPI operation indexes and
  emits OpenUdon UWS authoring/review/package artifacts.

Do not make OpenUdon import `../opentofu/internal/...` directly. Go's
`internal` visibility blocks that dependency shape, and the coupling would be
too brittle. Copy or adapt the needed static parser code into this repository
with explicit license/provenance tracking instead.

## License Rules

OpenTofu is licensed under MPL-2.0. Files copied from or adapted from
`../opentofu` remain MPL-2.0-covered.

When copying or adapting OpenTofu code:

- Preserve copyright and `SPDX-License-Identifier: MPL-2.0` headers.
- Do not relicense copied/adapted OpenTofu files as Apache-only or MIT-only.
- Keep copied/adapted OpenTofu code in clearly named package files.
- Update [UPSTREAM.md](UPSTREAM.md) with the OpenTofu commit, source paths, and
  local changes.
- Prefer small, focused copies over broad vendoring.
- Keep new, original glue files clearly separate when practical.

MPL-2.0 is file-level weak copyleft. Keep this repository's license boundary
obvious so downstream users can distinguish original `tfconfig` code from
OpenTofu-derived files.

## Product Scope

`tfconfig` is static analysis only.

In scope:

- Load a Terraform/OpenTofu configuration directory.
- Support `.tf`, `.tofu`, `.tf.json`, and `.tofu.json` files.
- Respect primary and override file behavior.
- Load local modules that are present on disk.
- Preserve variables, locals, outputs, providers, provider aliases, required
  providers, resources, data sources, dependencies, lifecycle, `count`,
  `for_each`, moved/import/removed/check/test facts, diagnostics, and source
  ranges.
- Emit deterministic Go structs and JSON.
- Preserve symbolic expressions and references without pretending to know
  runtime values.

Out of scope:

- Provider plugin execution.
- Provider schema loading in v1 unless explicitly designed later.
- `tofu init`, module download, backend initialization, state loading, refresh,
  planning, or apply.
- Credential resolution.
- OpenAPI operation mapping.
- UWS generation.
- OpenUdon review, package, approval, digest, or trusted-runner policy.

## Expected Output Contract

The target output is a stable static fact model, initially named
`tfconfig.static.v1`.

The model should make these distinctions explicit:

- literal values versus symbolic expressions;
- static references versus dynamic expressions;
- root module versus child modules;
- resources versus data sources;
- provider declarations versus provider uses;
- parser diagnostics versus conversion diagnostics.

OpenUdon conversion should be able to consume the JSON output without linking to
OpenTofu internals or executing Terraform/OpenTofu.

## Upstream Update Workflow

Use this workflow when updating code copied or adapted from `../opentofu`:

1. Check upstream state:

   ```bash
   git -C ../opentofu rev-parse HEAD
   git -C ../opentofu status --short
   ```

2. Identify the OpenTofu files relevant to static config parsing, usually under:

   ```text
   ../opentofu/internal/configs/
   ../opentofu/internal/configs/configload/
   ../opentofu/internal/command/jsonconfig/
   ```

3. Add or update explicit mappings in [sync/opentofu-files.tsv](sync/opentofu-files.tsv).
   Do not sync broad directories by default. Prefer syncing raw upstream
   snapshots under `_upstream/opentofu/...`, which the Go tool ignores, and
   then adapt the needed code into normal compile-ready `tfconfig` packages.

4. Compare the current `tfconfig` copy against those source files. Prefer
   deterministic commands such as `diff -u`, `git diff --no-index`, and `rg`.

5. Copy or adapt only the files needed for static parsing. For raw upstream
   snapshots, prefer:

   ```bash
   ./scripts/sync-opentofu.sh
   ```

   Do not import `../opentofu/internal/...` from this repo. Do not place raw,
   unadapted OpenTofu Go files in normal package directories if they import
   OpenTofu internals that have not been ported; use `_upstream/` for those
   snapshots.

6. Preserve MPL headers on copied/adapted files.

7. Remove or isolate behavior outside this repo's scope: CLI UI, provider
   plugins, backend/state/plan/apply behavior, cloud integrations, credential
   resolution, and non-static execution paths.

8. Update [UPSTREAM.md](UPSTREAM.md) with:

   - OpenTofu commit;
   - copied/adapted source paths;
   - local destination paths;
   - summary of local modifications;
   - known skipped upstream behavior.

   The sync script updates the managed provenance block automatically for
   allowlisted copies. Manually document any deeper adaptations that diverge
   from upstream.

9. Add or update fixtures before changing behavior where possible.

10. Run available checks:

   ```bash
   go test ./...
   go vet ./...
   git diff --check
   ```

If this repository is not yet a git repository, still run Go checks when a
module exists and use `git -C ../opentofu` only for upstream inspection.

## Fixture Expectations

Build fixtures around real Terraform/OpenTofu directory behavior:

- multiple `.tf` files in one module;
- `.tofu` files;
- `.tf.json` and `.tofu.json`;
- override files;
- variables and defaults;
- locals and references;
- provider aliases;
- required providers;
- resources and data sources;
- lifecycle;
- `depends_on`;
- `count`;
- `for_each`;
- local modules;
- missing modules;
- moved blocks;
- import blocks;
- removed blocks;
- check blocks;
- test files;
- invalid HCL diagnostics;
- deterministic source ranges.

Prefer small fixtures that isolate one behavior. Use broader integration
fixtures only after focused fixtures exist.

## Development Conventions

- Primary language is Go.
- Keep command entrypoints thin.
- Keep reusable parser/model logic under packages, not CLI code.
- Prefer structured parsing and OpenTofu-derived loader behavior over ad hoc
  string matching.
- Keep output deterministic: sort files, modules, blocks, diagnostics, and JSON
  arrays where the language does not define order.
- Do not execute side-effectful Terraform/OpenTofu commands from tests.
- Do not put secrets in fixtures. Use symbolic variables for credential-like
  values.

## Automation Guidance For Agents

Agents may update OpenTofu-derived code automatically when asked, but they must:

- Treat `../opentofu` as read-only unless the user explicitly asks to modify it.
- Never run `tofu init`, `tofu plan`, `tofu apply`, refresh, backend, or provider
  plugin commands as part of parser updates.
- Never silently drop upstream diagnostics or language constructs; preserve them
  or add explicit TODO diagnostics.
- Keep license headers intact.
- Update provenance in [UPSTREAM.md](UPSTREAM.md).
- Report which upstream commit and paths were used.
- Report which checks were run.

When in doubt, keep the parser conservative and review-first. The parser should
surface uncertainty as data or diagnostics rather than guessing Terraform
runtime behavior.

## GitHub Automation

The workflow [.github/workflows/sync-opentofu.yml](.github/workflows/sync-opentofu.yml)
runs on a weekly schedule and through manual dispatch. It checks out
`opentofu/opentofu`, runs [scripts/sync-opentofu.sh](scripts/sync-opentofu.sh),
runs Go checks, and opens a review-required pull request when allowlisted
OpenTofu files change.

Do not auto-merge sync pull requests. OpenTofu internals are public source, not
a stable public API.

# Architecture

## Repository Boundary

`tfconfig` owns a focused static parser boundary:

- selected OpenTofu source snapshots under `_upstream/opentofu/`;
- compile-ready parser/model/CLI code in normal Go packages;
- deterministic fixtures and optional upstream corpus tests;
- public `tfconfig.static.v1` Go structs and JSON projection.

Related repositories:

| Path | Role |
|---|---|
| `../opentofu` | Read-only upstream reference source. |
| `../tfconfig` | Static Terraform/OpenTofu parser package/tool and OpenTofu source mirror. |
| `../openudon` | Consumes `tfconfig` facts and OpenAPI indexes to produce reviewable UWS packages. |
| `../apitools` | OpenAPI loading, operation indexes, auth/security summaries, and ranking. |

## Static Data Flow

```text
Terraform/OpenTofu config directory
  -> tfconfig.LoadDir / LoadDirWithOptions
  -> tfconfig.static.v1 Document
  -> deterministic Go model and JSON projection
  -> OpenUdon conversion or fixture review
```

No Terraform/OpenTofu execution command is part of this flow.

## Parser Coverage

The model preserves:

- normalized source roots, files, and source ranges;
- root and local child modules with load status;
- variables, locals, outputs, providers, provider aliases, provider mappings,
  required providers, resources, data sources, and ephemeral resources;
- Terraform `backend`, `cloud`, and `provider_meta` blocks as static source
  facts only;
- module calls, lifecycle, dependencies, `count`, `for_each`, moved, import,
  removed, check, and test facts;
- schema-less provider/resource/data nested blocks as deterministic dotted
  config paths;
- symbolic expressions, references, parser diagnostics, and sensitive-candidate
  redaction metadata.
- object/map/tuple/list/set collection identity and top-level literal `toset`
  evaluation through a one-function whitelist; failed or context-dependent
  evaluation falls back to the original symbolic source fact.

## Boundary Rules

- Preserve OpenTofu-derived MPL-2.0 file provenance and headers.
- Do not import `../opentofu/internal/...`; copy or adapt required static
  parser code into this repository.
- Do not execute provider, backend, state, plan, apply, refresh, import, or
  credential behavior.
- Keep OpenAPI mapping, package artifacts, approval, digest, quality, and
  trusted-runner behavior in OpenUdon.

## Harness Layout

Private planning uses permanent `status-<LANE><NN>.md` ledgers. `M` retains
legacy and cross-cutting records, `P` owns static parser/model/fixture work,
and `U` owns OpenTofu mirror/provenance work. The roadmap is the only status
index; candidate directions remain unnumbered until promotion.

The unattended harness runner reads the second column of `Item | State |
Notes` tables. Cross-lane parallel work is valid only with explicit
non-overlapping file ownership, resolved prerequisites, and downstream impact.

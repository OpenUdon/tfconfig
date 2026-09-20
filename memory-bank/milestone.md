# Milestones

This file owns the `tfconfig` roadmap, lane meanings, active milestones,
status-file index, dependencies, candidate directions, scope, and acceptance
criteria. OpenUdon-owned conversion and package behavior lives in
`../openudon/memory-bank/milestone.md`; M07-M09 and M13-M14 remain here only as
permanent historical coordination records.

Per-task state lives in one `status-<LANE><NN>.md` file per milestone. Do not
copy task rows into this roadmap or create an aggregate `status.md`.

## Status ID Pattern

Status files are named `memory-bank/status-<LANE><NN>.md`, where `<LANE>` is
one uppercase domain letter and `<NN>` is a zero-padded number from `01`
through `99` within that lane.

Lane meanings:

- `M`: completed legacy work and future cross-cutting parser contracts.
- `P`: static parser, model, diagnostics, fixtures, and local-module behavior.
- `U`: OpenTofu mirror provenance, allowlist, sync automation, and upstream
  adaptation review.

Rules:

- Always use the two-digit form (`P01`, not `P1`) and never reuse an ID.
- M02-M09 were legacy identifiers. M15 explicitly normalizes their filenames
  and headings to M02-M09 without changing their historical identity. M01 was
  never allocated and remains unused; new work must not backfill it.
- Keep cancelled milestones and mark their rows `[X]` instead of removing or
  reusing them.
- Do not reclassify completed history merely to make lane sequences tidy.

## Parallel Work And Priority

Lane letters classify ownership, not execution order. Independent `P` and `U`
milestones may proceed together only when their sections record
non-overlapping file ownership, resolved prerequisites, and downstream impact.
Changes to shared public parser contracts stay in `M` or explicitly name every
cross-lane dependency. Prefer one active implementation milestone per lane.

## Current Dashboard

Active milestones: none. Latest completed milestone: P01 - Bounded Instance
Fact Evaluation.

## Status Files

Completed files remain permanent historical records.

| Milestone | Status File | State |
|---|---|---|
| M02 - Upstream Mirror Bootstrap And Provenance | [status-M02.md](status-M02.md) | Complete |
| M03 - `tfconfig.static.v1` Model Contract | [status-M03.md](status-M03.md) | Complete |
| M04 - Local Static Loader | [status-M04.md](status-M04.md) | Complete |
| M05 - Local Module Tree Loading | [status-M05.md](status-M05.md) | Complete |
| M06 - Parser Fixture Corpus | [status-M06.md](status-M06.md) | Complete |
| M07 - OpenUdon Convert TF Adapter Design | [status-M07.md](status-M07.md) | Complete |
| M08 - OpenUdon Convert TF Initial Implementation | [status-M08.md](status-M08.md) | Complete |
| M09 - Complete OpenUdon Package Integration | [status-M09.md](status-M09.md) | Complete |
| M10 - Parser Stewardship | [status-M10.md](status-M10.md) | Complete |
| M11 - OpenTofu Equivalence Fixture Hardening | [status-M11.md](status-M11.md) | Complete |
| M12 - OpenTofu Valid Modules Corpus Hardening | [status-M12.md](status-M12.md) | Complete |
| M13 - AWS Provider Single-OpenAPI Corpus | [status-M13.md](status-M13.md) | Complete |
| M14 - AWS Provider Multiple-OpenAPI Corpus | [status-M14.md](status-M14.md) | Complete |
| M15 - Parallel-Lane Harness Migration | [status-M15.md](status-M15.md) | Complete |
| P01 - Bounded Instance Fact Evaluation | [status-P01.md](status-P01.md) | Complete |

## Candidate Directions

Candidate directions are outside the active execution horizon. They have no
lane, permanent ID, status file, or execution-order entry until a fresh scope
and dependency review promotes them.

| Direction | Why Deferred | Promotion Trigger |
|---|---|---|
| Adapt another OpenTofu parser change | OpenTofu internals are not a stable API, and no bounded static-fact delta is currently approved. | An upstream change affects an allowlisted static parser fact and its provenance, compatibility, and fixture scope are reviewed. |
| Expand the static fact model | Extra fields increase the public contract and downstream compatibility burden without a current consumer requirement. | OpenUdon or another approved consumer demonstrates a provider-free static fact gap with fixtures. |
| Add another optional upstream corpus | Larger corpora add maintenance cost and can blur the no-execution boundary. | A parser fidelity risk cannot be covered adequately by focused local fixtures. |

## Review Procedure

When a milestone status file is completed:

1. Re-read its scope and acceptance criteria.
2. Review the relevant parser, model, fixture, documentation, or mirror diffs.
3. Confirm no provider, backend, state, plan, apply, module download, or
   credential behavior crossed into the static parser boundary.
4. Run `go test ./...`, `go vet ./...`, and `git diff --check`.
5. Update the memory bank and evolution record when contracts or direction
   change.

## Milestones

### M02 Upstream Mirror Bootstrap And Provenance

**Goal.** Establish the OpenTofu static parser reference mirror with clear
MPL-2.0 provenance and low-touch update automation.

**Acceptance.** The mirror is reproducible, tracks provenance and notices,
excludes runtime/planning code, and passes normal Go and diff checks.

### M03 `tfconfig.static.v1` Model Contract

**Goal.** Define the public static fact model that OpenUdon consumes.

**Acceptance.** The model represents multi-file configs, providers, aliases,
local modules, symbolic expressions, ranges, diagnostics, and safe JSON
without provider plugins, state, execution, or credentials.

### M04 Local Static Loader

**Goal.** Load real Terraform/OpenTofu configuration directories.

**Acceptance.** `.tf`, `.tofu`, HCL JSON, override, and test-file facts load
deterministically with parser diagnostics and no execution.

### M05 Local Module Tree Loading

**Goal.** Support local module trees without `tofu init` or remote downloads.

**Acceptance.** Direct local children load recursively; unavailable sources
become deterministic diagnostics; symbolic inputs and dependencies survive.

### M06 Parser Fixture Corpus

**Goal.** Lock parser behavior with focused fixtures.

**Acceptance.** Fixtures cover declarations, formats, modules, diagnostics,
ranges, determinism, and safe value projection.

### M07-M09 Historical OpenUdon Conversion Coordination

**Goal.** Record the completed downstream adapter design, initial conversion,
and package integration work that was originally coordinated from this
harness.

**Acceptance.** The permanent status files preserve the historical design and
implementation evidence. New converter behavior is planned in OpenUdon, not
in `tfconfig`.

### M10 Parser Stewardship

**Goal.** Keep the parser maintainable as OpenTofu and downstream conversion
evolve.

**Acceptance.** Sync, provenance, unsupported behavior, determinism,
redaction, module behavior, optional corpora, and downstream smoke checks stay
documented and provider-free.

### M11 OpenTofu Equivalence Fixture Hardening

**Goal.** Exercise the local equivalence corpus without provider execution.

**Acceptance.** Realistic provider-shaped configuration parses
deterministically and normal parser/downstream checks stay green.

### M12 OpenTofu Valid Modules Corpus Hardening

**Goal.** Use OpenTofu's static `valid-modules` corpus to close parser gaps.

**Acceptance.** High-value static constructs are represented without backend,
provider, test, plan, or apply execution.

### M13-M14 Historical AWS Conversion Corpus Coordination

**Goal.** Preserve the completed single- and multi-document AWS conversion
corpus records originally coordinated from this harness.

**Acceptance.** The status files remain historical evidence while future API
source selection and conversion behavior stay in apitools and OpenUdon.

### M15 Parallel-Lane Harness Migration

**Goal.** Adopt permanent lane-aware status IDs and parser-compatible task
ledgers without changing public parser behavior.

**Acceptance.** All historical files are indexed under canonical IDs, future
parser and upstream-mirror work has explicit ownership, candidates are
unnumbered, structural and repository checks pass, and the unattended runner
finds no actionable rows.

### P01 Bounded Instance Fact Evaluation

**Goal.** Expose the minimum deterministic static facts needed for downstream
count/for_each and local-module conversion without becoming an HCL runtime.

**Acceptance.** Collection shape distinguishes map/object/set from list/tuple,
literal numeric count and direct collections remain exact, a minimal pure
`toset` conversion can be recognized without variable context, symbolic
expressions stay symbolic, local-module call inputs/provider mappings remain
source-aware, JSON stays deterministic, and no module/provider/backend/state or
Terraform/OpenTofu execution is added.

**Completed.** P01 adds optional collection-shape identity, evaluates only a
top-level wholly known literal `toset`, strengthens instance/module fixtures,
and documents the public delta. Symbolic expressions remain symbolic and full
tfconfig, standalone, Ramen downstream, corpus, vet, and diff gates pass.

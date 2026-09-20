# Status P01 - Bounded Instance Fact Evaluation

State: Complete

## Goal

Give downstream conversion enough typed static evidence for bounded instance
and local-module expansion while preserving tfconfig's parser-only boundary.

## Task Status

| Item | State | Notes |
|---|---:|---|
| Scope and consumer evidence | `[+]` | Ramen M62 demonstrates that collection shape and a bounded pure set conversion are the missing facts; P01 excludes general evaluation and runtime context. |
| Collection-shape contract | `[+]` | tfconfig `da9f420` adds optional `collection_kind` with object/map/tuple/list/set identity, preserves it through safe literal projection and empty nested blocks, updates exact JSON fixtures, and tests deterministic source-aware output without changing redaction. |
| Bounded pure conversion | `[+]` | tfconfig `71b3840` evaluates only a top-level whitelisted literal `toset`, preserves its deterministic set shape and de-duplicated values, and proves variable-backed `toset` plus unknown functions remain source-aware symbolic expressions. |
| Module and instance fixtures | `[+]` | tfconfig `55f0d4d` strengthens local/nested module fixtures for numeric count, set values, full addresses, inputs, provider mappings, and source ranges; the focused collection tests prove object keys, exact JSON, and symbolic rejection, and the public static-v1 contract documents the whitelist. |
| Downstream qualification and review | `[+]` | Full and standalone tfconfig tests/vet plus Ramen focused/full tests, vet, exact corpus check, and diff checks pass; review confirms the optional field and one-function whitelist add no variable/local resolution, provider/backend/state/module download, network, or execution behavior. |

## Scoped Commits

- Tofu lane opening: `3cca6d0`
- tfconfig collection shape: `da9f420`
- tfconfig bounded `toset`: `71b3840`
- tfconfig fixtures/docs: `55f0d4d`
- Tofu closure: this commit

## Boundaries

- No general Terraform/OpenTofu expression evaluator.
- No variable/local resolution, provider functions, filesystem functions,
  environment reads, time/randomness, network, module download, state, plan,
  apply, or provider loading.
- `hcllight` remains a useful generic-HCL reference but is not a dependency;
  tfconfig owns source-aware Terraform/OpenTofu fact projection.

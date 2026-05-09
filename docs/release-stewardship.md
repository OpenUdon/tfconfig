# Parser Stewardship

This checklist keeps `tfconfig` static, deterministic, and provenance-safe as
OpenTofu and downstream OpenUdon conversion evolve.

## OpenTofu Sync Review

- Inspect `sync/opentofu-files.tsv` for intentionally mirrored source files.
- Use `OPENTOFU_DIR=../opentofu ./scripts/sync-opentofu.sh` to refresh raw
  snapshots from a local upstream checkout.
- Review `_upstream/opentofu/...` as reference material only; compile-ready
  behavior must be adapted into normal `tfconfig` packages.
- Update `UPSTREAM.md`, `THIRD_PARTY.md`, and MPL license records when copied
  or adapted files change.
- Preserve copyright and `SPDX-License-Identifier: MPL-2.0` headers on
  OpenTofu-derived files.

## Static Boundary Review

Parser changes must remain static. Treat these as out of scope:

- provider plugin execution or provider schema loading;
- `tofu init`, remote module download, backend initialization, state, refresh,
  plan, apply, import, or test execution;
- credential resolution or secret storage;
- OpenAPI mapping, UWS generation, approval, digest, quality, handoff, or
  trusted execution.

## Regression Matrix

| Area | Primary checks |
|---|---|
| Deterministic JSON/model output | `model_test.go`, `load_test.go`, fixture corpus |
| Safe value projection | sensitive value model/load tests and fixture leak checks |
| Local and missing modules | local module tree, unavailable source, unreadable source, cycle guard, and module manifest tests |
| OpenTofu equivalence corpus | `TestOpenTofuEquivalenceFixtureCorpus` when sibling checkout exists |
| OpenTofu valid-modules corpus | `TestOpenTofuValidModulesFixtureCorpus` when sibling checkout exists |
| Downstream smoke | `(cd ../openudon && go test ./internal/tfconvert)` |

## Normal Gates

```bash
go test ./...
go vet ./...
git diff --check
```

Before a public OpenUdon release that depends on new parser behavior, publish a
`tfconfig` revision, update OpenUdon's module version, and verify OpenUdon with
`GOWORK=off`.

# tfconfig Boundary

`github.com/OpenUdon/tfconfig` is the parser package/tool for this conversion
work. It exists because importing OpenTofu internals directly is not a viable
boundary and because OpenUdon should not absorb Terraform/OpenTofu runtime
semantics.

## Source Mirror

`../tfconfig` mirrors selected OpenTofu source under:

```text
../tfconfig/_upstream/opentofu/
```

The mirror is maintained by:

```text
../tfconfig/sync/opentofu-files.tsv
../tfconfig/scripts/sync-opentofu.sh
../tfconfig/.github/workflows/sync-opentofu.yml
```

It currently excludes OpenTofu `internal/tofu`, which is runtime/planning code.
Ongoing sync review and release checks are tracked in
[release-stewardship.md](release-stewardship.md).

## License Boundary

OpenTofu is MPL-2.0. OpenTofu-derived files copied or adapted into `tfconfig`
remain MPL-2.0-covered. Provenance and notices live in:

```text
../tfconfig/UPSTREAM.md
../tfconfig/THIRD_PARTY.md
../tfconfig/licenses/opentofu-MPL-2.0.txt
```

## Implementation Boundary

Raw upstream snapshots under `_upstream/` are reference material. Compile-ready
`tfconfig` code should adapt only what is needed into smaller static packages.

`tfconfig` should expose a stable public model and avoid leaking OpenTofu
internal package shapes into OpenUdon.

OpenUdon may import `github.com/OpenUdon/tfconfig`, but not
`github.com/opentofu/opentofu`, `github.com/hashicorp/terraform`, or
`github.com/OpenUdon/tfconfig/_upstream/...`. The
`openudon check-apitools-boundary` gate checks this repository boundary along
with the narrowed `apitools` API boundary.

## v1 Static Policy

Allowed:

- read local config directories;
- read local module directories;
- parse HCL and HCL JSON;
- preserve symbolic expressions, references, and diagnostics.

Rejected or diagnostic-only:

- provider plugins;
- remote module downloads;
- backend/state;
- refresh/plan/apply;
- credential resolution;
- execution behavior.

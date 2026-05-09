# OpenTofu Sync Research

Date: 2026-05-09

This note maps the `../tofu/ideas.md` conversion requirements to the OpenTofu
internal files that should be mirrored into `tfconfig` for reference.

## Requirement Summary

`tfconfig.static.v1` needs static Terraform/OpenTofu configuration facts:

- directory-based multi-file loading;
- `.tf`, `.tofu`, `.tf.json`, and `.tofu.json`;
- override files;
- local modules and static module trees;
- variables, locals, outputs;
- provider configs, aliases, and required providers;
- resources and data sources;
- lifecycle, `depends_on`, `count`, and `for_each`;
- moved, import, removed, check, and test blocks;
- diagnostics and source ranges;
- no provider plugins, backend/state/plan/apply, refresh, or credential
  resolution.

## First Mirror Set

The first sync allowlist mirrors hand-written OpenTofu source files under
`_upstream/opentofu/...`. That directory is ignored by the Go tool, so raw
OpenTofu internals can be tracked without making this module compile them.

The active allowlist in [../sync/opentofu-files.tsv](../sync/opentofu-files.tsv)
contains:

- `internal/configs/*.go` hand-written source needed for static config parsing,
  decoding, merge, validation, and static module tree behavior;
- `internal/configs/configload/*.go` hand-written source for loader/module
  manifest behavior;
- `internal/command/jsonconfig/*.go` hand-written source for optional JSON
  projection reference.

This is the correct first mirror because OpenTofu's config behavior is cohesive:
file selection, block decoding, module merge, provider requirements, references,
and diagnostics are spread across the package rather than isolated in only
`parser.go`.

## Important Files

Highest-value files for the first `tfconfig` port:

- `internal/configs/parser.go`: parser cache and HCL/JSON syntax selection.
- `internal/configs/parser_config_dir.go`: directory file discovery, `.tf` vs
  `.tofu`, JSON variants, overrides, and test file selection.
- `internal/configs/parser_config.go`: top-level config/test file decode
  dispatch.
- `internal/configs/module.go`: per-directory module container and file merge
  entrypoint.
- `internal/configs/module_merge.go`: override and duplicate handling.
- `internal/configs/config_build.go`: static module tree construction.
- `internal/configs/resource.go`: resource/data/lifecycle/count/for_each
  decoding.
- `internal/configs/named_values.go`: variables, locals, outputs, type/default
  behavior.
- `internal/configs/provider.go`: provider config and alias decoding.
- `internal/configs/provider_requirements.go`: required provider facts.
- `internal/configs/module_call.go`: module source/count/for_each/providers and
  dependency handling.
- `internal/configs/depends_on.go`: dependency traversal decoding.
- `internal/configs/moved.go`, `import.go`, `removed.go`, `checks.go`,
  `test_file.go`: review-relevant facts and diagnostics.
- `internal/configs/configload/loader_load.go`: OpenTofu's installed-module
  loader behavior, useful as a contrast for `tfconfig`'s local-static v1
  behavior.
- `internal/command/jsonconfig/expression.go`: expression/reference JSON shape
  reference, useful if `tfconfig` exposes expression facts similar to
  OpenTofu's JSON.

## Deferred Generated Files

The first allowlist does not mirror these generated files because they lack
MPL-2.0 SPDX headers in the local OpenTofu checkout:

- `internal/configs/provisioneronfailure_string.go`
- `internal/configs/provisionerwhen_string.go`
- `internal/configs/variabletypehint_string.go`

If the port needs equivalent string methods, regenerate them in `tfconfig` or
copy them only after deciding how to preserve license provenance for generated
OpenTofu-derived files.

## Deferred Dependency Packages

The active sync does not yet mirror every package imported by OpenTofu
`internal/configs`, such as:

- `internal/addrs`
- `internal/configs/configschema`
- `internal/configs/hcl2shim`
- `internal/tfdiags`
- `internal/lang`
- `internal/getmodules`
- `internal/getproviders`
- `internal/instances`
- `internal/depsfile`
- `internal/encryption/config`

For `tfconfig`, these should be handled case by case. Some concepts should be
ported into small public/static equivalents, while others should be cut because
they belong to provider execution, backend/state behavior, or OpenTofu runtime
semantics.

## Deferred Tests And Fixtures

Do not bulk-copy all upstream testdata yet. Start with focused fixtures based on
the v1 scope:

- `internal/configs/testdata/tofu-and-tf-files`
- `internal/configs/testdata/tofu-only-files`
- `internal/configs/testdata/valid-files`
- `internal/configs/testdata/invalid-files`
- `internal/configs/testdata/config-build`
- `internal/configs/testdata/nested-errors`
- `internal/configs/testdata/uninit-module-and-provider-refs`
- `internal/configs/configload/testdata/local-modules`
- `internal/configs/configload/testdata/already-installed`
- `internal/configs/configload/testdata/already-installed-now-invalid`
- `internal/command/testdata/show-config-module`
- `internal/command/testdata/show-config-single-module`

Mirror fixtures only when a matching `tfconfig` test is being added. This keeps
the repository smaller and avoids inheriting behavior that v1 intentionally
does not support.

## Porting Guidance

Use the mirrored `_upstream` files as reference, then adapt into normal
compile-ready packages with a smaller public model. Do not try to compile raw
OpenTofu internals in place.

Likely `tfconfig` package split:

- root package: public `LoadDir` API and `tfconfig.static.v1` model;
- internal loader package: file discovery and parse orchestration;
- internal expression package: symbolic expression/reference extraction;
- cmd package later: JSON CLI wrapper.

Keep v1 local-static:

- load local modules already present on disk;
- report missing/remote modules as diagnostics;
- preserve expressions symbolically;
- never run `tofu init`, provider plugins, backend/state/plan/apply, or refresh.

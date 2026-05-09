# tfconfig.static.v1

`tfconfig.static.v1` is the initial public fact model emitted by
`github.com/OpenUdon/tfconfig` and consumed by OpenUdon conversion.

The contract is static and review-first. It records Terraform/OpenTofu source
facts, symbolic expressions, references, source ranges, diagnostics, and module
structure without provider plugin execution, `tofu init`, module downloads,
backend or state loading, refresh, plan, apply, or credential resolution.

## Public Boundary

The primary integration boundary is the Go API in package `tfconfig`.
Deterministic JSON is an export/debug projection of the same model and is safe
to write to review artifacts.

`LoadDir` is the first local-static loader entrypoint:

```go
doc, err := tfconfig.LoadDir("./tf")
```

It discovers `.tf`, `.tofu`, `.tf.json`, `.tofu.json`, override files, and
module test files in the root module directory. Direct local child module
sources such as `./modules/app`, `../shared`, and absolute filesystem paths are
loaded recursively when the directories are readable. Registry, Git, HTTP, S3,
OCI, symbolic, missing, and other downloader-backed module sources are not
fetched; they are represented as child modules with load status and
diagnostics. `.terraform/modules/modules.json` is not read by the static
loader.

The top-level document uses:

```json
{
  "version": "tfconfig.static.v1",
  "producer": "github.com/OpenUdon/tfconfig",
  "root_dir": "./tf",
  "source_files": [],
  "modules": [],
  "diagnostics": []
}
```

## Model Coverage

The model includes:

- normalized source roots, files, and source ranges;
- root and child modules with load status;
- variables, locals, outputs, defaults, expressions, references, and sensitivity
  markers;
- required providers, provider configs, aliases, provider references, and
  provider mappings;
- managed resources and data sources with attributes, lifecycle,
  `depends_on`, `count`, `for_each`, references, and source ranges;
- module calls with source, inputs, provider mappings, dependencies, `count`,
  and `for_each`;
- moved, import, removed, check, and test facts;
- parser diagnostics distinct from later OpenUdon conversion diagnostics.

Provider, managed resource, and data source bodies are treated as schema-less
provider configuration surfaces. Unknown nested provider blocks are preserved as
dotted config attribute paths rather than rejected, because `tfconfig` does not
execute provider plugins or load provider schemas. Repeated nested blocks use
source-order indexes, for example `first_block[0].id` and `first_block[1].id`.
The built-in Terraform/OpenTofu meta blocks that `tfconfig` models directly,
such as resource `lifecycle`, remain represented in their explicit model fields.

## Value Projection

Values distinguish literals, symbolic expressions, unknowns, collections, and
redacted sensitive or likely-secret literals.

Symbolic references such as `var.api_token` remain expressions with references.
Values marked as likely secret candidates are emitted as:

```json
{
  "kind": "redacted",
  "sensitive_candidate": {
    "reason": "attribute name suggests secret material",
    "attribute_path": "token"
  },
  "redacted": true
}
```

The public JSON projection does not emit raw literals for values marked as
sensitive, redacted, or as sensitive candidates.

## Determinism

`Document.JSON` and `Document.JSONIndent` return deterministic public JSON.
Before encoding, the model is canonicalized by sorting source roots, source
files, modules, module-local declarations, resources, data sources, module
calls, references, diagnostics, and repeated structural facts.

This deterministic projection is intended for fixtures, review diffs, and the
future `tfconfig` CLI JSON output.

# Status M13 — AWS Provider Single-OpenAPI Corpus

State of M13 items. See [milestone.md](milestone.md) for milestone scope and
acceptance criteria.

Status markers:

| Symbol | Suggested Status | Interpretation |
|---|---|---|
| `[ ]` | Pending | Item not started or pending action. |
| `[+]` | Completed | Item finished or done. |
| `[~]` | In Progress | Item is being worked on. |
| `[!]` | Blocked | Item requires attention or is on hold. |
| `[X]` | Cancelled | Item is no longer needed. |

Each table row is a commit unit once implementation begins: after flipping a
row to `[+]`, verify the change, update the memory bank/docs, and make a scoped
commit before starting the next row. If multiple rows are inseparable, use one
coherent commit and name every covered row in the handoff.

| Item | State | Notes |
|---|---|---|
| AWS S3 provider tests selected | `[+]` | Initial corpus uses `TestAccS3BucketAccelerateConfiguration_basic` and `TestAccS3BucketDataSource_basic` from `../terraform-provider-aws/internal/service/s3` |
| Single OpenAPI document imported | `[+]` | `../apitools` imports APIs.guru S3 `2006-03-01` with 97 operations |
| Static conversion smoke run | `[+]` | Non-strict `openudon convert tf` produced package artifacts for both S3 cases without Terraform/OpenTofu/provider/AWS execution |
| First strict failures registered | `[+]` | Strict mode fails on ambiguous/unresolved S3 operation matching; see `docs/aws-provider-conversion-corpus.md` |
| Reproducible fixture corpus added | `[+]` | `../openudon/internal/tfconvert:TestConvertAWSProviderS3SingleOpenAPICorpus` covers the S3 bucket acceleration and S3 bucket data-source snippets with a stable local S3 OpenAPI fixture |
| AWS operation ranking improved | `[+]` | OpenUdon now maps `aws_s3_bucket` create to `CreateBucket`, `aws_s3_bucket_accelerate_configuration` create to `PutBucketAccelerateConfiguration`, and `data.aws_s3_bucket` read to `GetBucketLocation` before generic text ranking |
| Quality gate expectations tightened | `[+]` | The S3 corpus test asserts no operation TODOs, no `operation.ambiguous`/`operation.unresolved` diagnostics, and no quality failures for `intent.openapi_operations` or `conversion.diagnostics` |

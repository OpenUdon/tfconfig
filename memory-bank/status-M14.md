# Status M14 — AWS Provider Multiple-OpenAPI Corpus

State of M14 items. See [milestone.md](milestone.md) for milestone scope and
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
| AWS multi-service provider test selected | `[+]` | Initial corpus uses `TestAccLambdaFunctionURL_basic` from `../terraform-provider-aws/internal/service/lambda/function_url_test.go` |
| Multiple OpenAPI documents located | `[+]` | APIs.guru IAM `2010-05-08`, Lambda `2015-03-31`, and STS `2011-06-15` documents exist and were downloaded for the smoke pass |
| Static conversion smoke run | `[+]` | Non-strict `openudon convert tf` produced package artifacts for the Lambda/IAM/STS case without Terraform/OpenTofu/provider/AWS execution |
| First strict failures registered | `[+]` | Initial strict failures for IAM/Lambda operation matching, partition metadata, SigV4 credential binding, and Lambda URL path-parameter binding were registered before the M14 fixes; see `docs/aws-provider-conversion-corpus.md` |
| apitools AWS import failure registered | `[+]` | `../apitools` now accepts the IAM, Lambda, and STS raw APIs.guru documents through a narrowed AWS OpenAPI download fallback while preserving external-reference rejection |
| Reproducible multi-document fixture added | `[+]` | OpenUdon now has committed `TestConvertAWSProviderLambdaFunctionURLMultiOpenAPICorpus` coverage with local IAM, Lambda, and STS OpenAPI fixtures |
| AWS service routing improved | `[+]` | AWS IAM and Lambda resources now map to deterministic service operation IDs such as `POST_CreateRole`, `POST_PutRolePolicy`, `CreateFunction`, and `CreateFunctionUrlConfig` |
| Provider-local AWS data sources classified | `[+]` | `data.aws_partition.current` and related provider-local AWS metadata are preserved symbolically; `data.aws_caller_identity` remains an STS API read |
| AWS request binding improved | `[+]` | Terraform `function_name` now binds the Lambda Function URL `FunctionName` path parameter, and AWS query-protocol `Action`/`Version` constants are bound for IAM/STS operations |
| SigV4 credential binding modeled | `[+]` | AWS `hmac`/SigV4 security now creates auditable symbolic credential bindings, preserving provider aliases such as `aws_west_hmac`, without resolving credentials |

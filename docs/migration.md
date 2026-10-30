# Migration from the AWS upstream

This tree started from `airbnb/rudolph` at `34f5ebb` (the `master` commit that skips empty Kinesis `PutRecords` calls). The fork was empty apart from a README, so that commit was imported and then the AWS implementation was removed. There is no remaining AWS SDK dependency and no AWS Terraform.

The Santa wire protocol is the same. A client enrolled against this server uses the same preflight, ruledownload, eventupload, and postflight bodies and the same cursor JSON (`strategy`, `pk`, `sk`, `page`, `batch_size`). Rule types `BINARY`, `CERTIFICATE`, `SIGNINGID`, and `TEAMID` and the policy names are unchanged. Clean sync versus feed sync uses the same decision.

## What moved

| Upstream | This fork |
| --- | --- |
| API Gateway REST API, one Lambda per route, shared `cmd/api` zip | One Cloud Run container. The same router still dispatches on the route template |
| Custom REQUEST authorizer that allowed every POST with a `machine_id` | mTLS at the external HTTPS load balancer. The authorizer binary is gone |
| Optional API Gateway resource policy of source IPs | Optional Cloud Armor policy |
| DynamoDB table `{prefix}_rudolph_store` with PK/SK and TTL on `ExpiresAfter` | Firestore Native database, same PK/SK fields, TTL on `expireAt` |
| Kinesis, Firehose-to-S3, and an extra Lambda invoke | One Pub/Sub publish. A Cloud Storage subscription archives objects. Add subscriptions for anything that used to be the extra Lambda |
| S3 bucket for Lambda zips | Artifact Registry |
| KMS keys on the table and buckets | Cloud KMS key on the events bucket |
| CloudWatch log groups | Cloud Logging via Cloud Run stdout |
| Operator IAM against DynamoDB | Application Default Credentials against Firestore |
| `REGION`, `DYNAMODB_NAME`, `HANDLER`, `KINESIS_NAME`, `FIREHOSE_NAME`, `LAMBDA_NAME` | `GCP_PROJECT`, `FIRESTORE_DATABASE`, `EVENT_TOPIC`, `REQUIRE_CLIENT_CERT`, `PORT` |

## Behaviour that is not identical

These are the gaps to plan for. They do not change a successful Santa sync.

1. **Authentication is stricter.** Upstream allowed any client that could reach API Gateway and POST a UUID. This server requires a certificate signed by `client_ca_pem`. Existing Santa profiles that only set `SyncBaseURL` will fail the TLS handshake until MDM installs a client identity. See [certificates.md](certificates.md).
2. **Firestore IAM cannot copy the DynamoDB key-prefix condition.** Upstream sync Lambdas could write only `Machine#`, `MachineInfo#`, and `MachineRules#` items. The Cloud Run service account can write the whole database. The HTTP API still has no route that creates global rules. Do not distribute the runtime service account key; run the CLI as yourself.
3. **There is no direct function invoke on the event path.** Upstream could set `LAMBDA_NAME` and asynchronously invoke that function in addition to Kinesis or Firehose. Publish once to Pub/Sub and attach a push subscription if you still need an HTTP consumer. The JSON body is the same forwarded event (`machine_id` plus the Santa event fields).
4. **Event export is on by default in Terraform** (`enable_event_export = true`) because an archive is the useful replacement for Firehose. Set it to false to restore the old default of accepting events and dropping them. An empty event list still does not call the publisher.
5. **`lookup machine-ids` has an index.** Upstream's CLI queried a GSI named `DataType_MachineID` that the Terraform did not create. This module creates the collection-group index, so a prefix search works on a fresh database.
6. **The 10 minute feed-cursor rewind is unchanged.** It was a hedge against DynamoDB's eventually consistent queries. Firestore point reads and the queries used here are strongly consistent. The rewind still causes a small overscan and is preserved so two overlapping syncs cannot skip a feed write.
7. **Pagination cursors are exclusive on SK.** The wire format is the same. A full page still returns a cursor, and the following call can be an empty page, as with DynamoDB's `LastEvaluatedKey` when the page size lands exactly on the end.
8. **Numeric rule and policy values are stored as integers**, not DynamoDB numeric strings. Reads still accept the old numeric strings (`"1"` … `"6"`) and the legacy data-type aliases, which matters only if you import documents that were exported from DynamoDB.
9. **`/logupload` is still not implemented.** `/xsrf` is still a stub. The repair command still scans for legacy `MachineInfo#` and `MachineConfig#` keys and still does not delete them unless you change the hardcoded flag in the command, which is how upstream shipped it.
10. **Usage plans and API keys are gone.** Nothing in the Santa protocol used them. The authorizer's `UsageIdentifierKey` has no equivalent.
11. **There is no data migration tool in this repo.** Standing up a new fleet is a new Firestore database and a new `SyncBaseURL`. Moving an existing DynamoDB table means exporting items (PK, SK, and attributes) and writing the same fields into `partitions/{base64url(PK)}/items/{base64url(SK)}`, with `expireAt` set from `ExpiresAfter` when that attribute is a positive unix time. Rule type and policy numbers should be integers. Do that with a one-off job under a service account that can read DynamoDB and write Firestore; this repository does not ship that job because the AWS SDK has been removed.

## CLI configuration

Replace the old tfvars keys the CLI understood (`region`, `aws_account_id`, `prefix` used to build a table name) with:

```json
{
  "prefix": "prod",
  "project_id": "your-project",
  "firestore_database": "rudolph",
  "region": "us-central1"
}
```

Flags are `--project` and `--firestore_database`. `--region` and `--dynamodb_table` are gone.

## Rollout order

1. Apply the stack and wait until the managed certificate is active.
2. Issue a client certificate for a pilot Mac and install it with the Santa profile.
3. Confirm a full sync in Cloud Logging: preflight, one or more ruledownload calls, eventupload, postflight.
4. Load global rules with the CLI before you expect the pilot to receive anything but the default config.
5. Move the rest of the fleet only after the pilot syncs on a schedule. Clients do not need to know that the server used to be Lambda.

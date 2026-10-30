# Architecture

Rudolph on Google Cloud is one stateless HTTP service plus a document database. Santa never talks to Cloud Run directly.

## Request path

1. Santa presents a client certificate signed by the organisation CA.
2. A global external Application Load Balancer terminates TLS. A Certificate Manager trust config holds that CA, and a Network Security server TLS policy sets `client_validation_mode` to `REJECT_INVALID`. Requests without a trusted client certificate never reach the application.
3. The load balancer adds request headers describing the client certificate (`X-Client-Cert-Chain-Verified`, subject, issuer, fingerprint, serial). When `REQUIRE_CLIENT_CERT=true`, the API rejects non-health calls whose chain was not verified. `/health` stays exempt so a process-local probe does not need a certificate. External callers of `/health` still pass through the load balancer, so they need a client certificate too.
4. Cloud Run ingress is `INGRESS_TRAFFIC_INTERNAL_LOAD_BALANCER`. Combined with `invoker_iam_disabled`, the only network path onto the service is the load balancer. The load balancer does not send a Cloud Run identity token, so IAM invoker checks are not the security boundary.
5. Optional Cloud Armor source CIDRs sit in front of mTLS when `allowed_inbound_cidrs` is set. That replaces the old API Gateway resource policy.

The process is `cmd/api`. It listens on `$PORT` (Cloud Run sets this) and routes:

| Method | Path | Role |
| --- | --- | --- |
| GET | `/health` | Firestore read of one global-rule page |
| POST | `/preflight/{machine_id}` | Store sensor data and sync state, return config |
| POST | `/ruledownload/{machine_id}` | Page global rules, the rules feed, then machine rules |
| POST | `/eventupload/{machine_id}` | Publish execution events, or accept and drop them |
| POST | `/postflight/{machine_id}` | Finish the sync cursor and delete stale machine rules |
| POST | `/xsrf/{machine_id}` | Same stub response as upstream |

`machine_id` must be a UUID. Bodies are JSON. Rule download cursors are still an opaque JSON string that Santa echoes back.

## Firestore

The logical model is still a partition key (`PK`) and a sort key (`SK`).

| Data | PK | SK |
| --- | --- | --- |
| Global rule | `GlobalRules` | `Binary#`, `Cert#`, `TeamID#`, or `SigningID#` plus the identifier |
| Rules feed | `RulesFeed` | `{RFC3339}#{rule sort key}` |
| Machine rule | `MachineRules#{uuid}` | rule sort key |
| Global config | `GlobalConfig` | `Config` |
| Machine config | `Machine#{uuid}` | `Config` |
| Sync state | `Machine#{uuid}` | `SyncState` |
| Latest sensor snapshot | `Machine#{uuid}` | `Current` |

Each item is a document at `partitions/{base64url(PK)}/items/{base64url(SK)}`. Queries order by the `SK` field, not by the document id, so feed pagination stays lexicographic. A cursor's sort key is exclusive: the next page starts at the first SK strictly greater than the cursor. A full page returns a cursor even when it might be the last page, which matches the extra round trip DynamoDB produced.

`ExpiresAfter` remains a unix timestamp on the document. The client also writes `expireAt` as a timestamp, and Terraform installs a TTL policy on `expireAt` for the `items` collection group.

Indexes created with the database:

- Collection group `SerialNum` + `DataType` for `rudolph lookup serial-number`.
- Collection group `DataType` + `MachineID` for `rudolph lookup machine-ids`. Upstream referenced this index from the CLI but did not create it. It is created here.
- Collection group `PK` + `SK` for the repair scan.
- Collection `DeleteOnNextSync` + `SK` for postflight deletion of machine rules.

Point-in-time recovery is enabled on the database.

Global rule changes are a transaction that updates `GlobalRules` and appends a `RulesFeed` document, including an explicit `REMOVE` feed entry when a rule is deleted. The feed cursor stored at postflight is still "now minus 10 minutes". Firestore reads are strongly consistent, so that rewind is no longer required for the old DynamoDB race, but it is kept so a client cannot skip a feed entry that was written while its sync was already in flight.

## Events

`EVENT_TOPIC` empty means event upload returns `{"status":"ok"}` and stores nothing, which is the old `HANDLER=NONE` behaviour. An empty batch does not call Pub/Sub.

When the topic is set, each Santa event is published as JSON with the machine id both in the body and as a message attribute and ordering key. Terraform attaches a Cloud Storage subscription that writes newline-batched objects under `santa-events/`. That subscription is the Firehose replacement. Extra processing is another subscription on the same topic, which replaces invoking a second Lambda.

The events bucket uses a Cloud KMS key, uniform bucket-level access, public access prevention, and a lifecycle rule.

## Identity

The Cloud Run service account can read and write the Firestore database, publish to the events topic, and write logs. Firestore IAM does not support the DynamoDB-style condition that limited the sync role to `Machine#` and `MachineRules#` key prefixes. The public API still has no route that writes global rules or the feed; only the CLI does. Treat the runtime service account as privileged and do not hand out its key. Operators run the CLI with their own credentials.

The client CA PEM is also stored in Secret Manager. The load balancer trust config is what actually authenticates clients; the secret is the copy operators can retrieve later.

## Build and logs

Images live in Artifact Registry. Cloud Run sends stdout and stderr to Cloud Logging. The server logs the method, route template, and status code, not the request body.

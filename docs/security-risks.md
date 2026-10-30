# Security

## Who can call the API

The public hostname requires a client certificate that chains to the CA in the Certificate Manager trust config. The server TLS policy rejects invalid or missing certificates before Cloud Run runs. Cloud Run ingress is limited to the load balancer, so the `*.run.app` URL is not a second front door.

That is a different threat model from upstream Rudolph, where API Gateway accepted any POST that looked like a sync. Here, possession of a trusted client certificate is the admission check. The machine id in the path is not independently bound to the certificate subject. A stolen client certificate can sync as any machine id the caller chooses, and can download the rule set. Protect the device identity the same way you protect other MDM credentials, and prefer one certificate per Mac so you can remove a single identity.

`REQUIRE_CLIENT_CERT` on the container checks the `X-Client-Cert-Chain-Verified` header the load balancer sets. It is a second check, not a replacement for the TLS policy. `/health` does not require the header. Callers on the public hostname still need a certificate to reach `/health`, because the load balancer enforces mTLS for every request.

Optional `allowed_inbound_cidrs` adds a Cloud Armor allow list. Leave it empty to allow any network that presents a valid certificate.

## Who can change rules

The sync routes do not create or edit global rules. Rule and configuration writes go through the CLI, using the caller's Google credentials. The Cloud Run service account is broader than the old Lambda role: Firestore IAM cannot limit it to `Machine#` and `MachineRules#` documents. Do not export a key for that service account. Grant operators `roles/datastore.user` (or a custom role) on the project instead of sharing the runtime account.

## Server identity

The server certificate is a Google-managed certificate for `domain_name`. Santa trusts public CAs already in the macOS trust store. If you previously pinned Amazon's certificate chain because of the upstream note about `ServerAuthRootsFile`, re-read that Santa setting against the version you ship. A managed Google certificate does not use an AWS CA.

## Data

Firestore point-in-time recovery is on. The events bucket blocks public access, uses uniform bucket-level access, and is encrypted with a Cloud KMS key. Secret Manager holds a copy of the client CA certificate. That object is not a private key. The CA private key should never be in this project.

Logs record the route and status, not the sync body. Event payloads are in Pub/Sub and the archive bucket, which is where execution details (paths, hashes, signing chains) live. Restrict both to the people who already handle Santa telemetry.

# Rudolph

Rudolph is a sync server for [Santa](https://github.com/google/santa), the macOS binary authorization agent. This fork runs only on Google Cloud.

Santa clients sync through an external Application Load Balancer that requires a client certificate. The load balancer forwards those requests to a single Cloud Run service. Rules, machine configuration, sync cursors, and the latest preflight snapshot live in a Firestore Native database. Execution events can be published to Pub/Sub and archived in Cloud Storage.

The Santa sync protocol is unchanged: `POST /preflight/{machine_id}`, `/ruledownload/{machine_id}`, `/eventupload/{machine_id}`, and `/postflight/{machine_id}`, plus `GET /health`. Operators still change rules and configuration with the `rudolph` CLI, not through the public API.

## Architecture

```text
Santa (client cert) --> external HTTPS LB (mTLS, Certificate Manager trust config)
                    --> Cloud Run (ingress: internal + load balancer only)
                    --> Firestore Native
                    --> Pub/Sub --> Cloud Storage   (optional event archive)
```

| Former AWS piece | GCP replacement |
| --- | --- |
| API Gateway + Lambda | Cloud Run service |
| Custom authorizer | mTLS on the external HTTPS load balancer |
| DynamoDB | Firestore Native |
| Kinesis / Firehose / Lambda fan-out | Pub/Sub topic and a Cloud Storage subscription |
| S3 (artifacts and event archive) | Artifact Registry and Cloud Storage |
| KMS | Cloud KMS key used for the events bucket |
| IAM roles | Service account on the Cloud Run service |
| CloudWatch Logs | Cloud Logging (stdout from Cloud Run) |
| Secrets Manager / SSM | Secret Manager (client CA certificate) |
| Route 53 | Optional Cloud DNS records |
| Terraform AWS provider | Terraform Google provider |

Details, including behaviour that is intentionally unchanged and behaviour that is not identical, are in [docs/architecture.md](docs/architecture.md) and [docs/migration.md](docs/migration.md).

## Layout

- `cmd/api` is the Cloud Run HTTP server.
- `cmd/cli` is the operator CLI.
- `pkg/store` is the Firestore persistence layer. An in-memory implementation backs the unit tests.
- `deployments/terraform_modules/santa_api` is the Google Cloud stack.
- `deployments/environments/example` is a root module you can copy.

## Build and test

Go 1.22 or newer:

```bash
make test
make build
```

`make build` writes `build/rudolph-api` and `build/rudolph`. A container image is:

```bash
docker build -t rudolph-api:local .
```

Push that image to the Artifact Registry repository Terraform creates, then set `container_image` to the tagged URI and apply again. Cloud Run will not start until the image exists.

## Deploy

1. Pick a GCP project and a DNS name, for example `sync.example.com`.
2. Copy `deployments/environments/example` or run `ENV=prod make init-env`.
3. Edit `config.auto.tfvars.json`: project, region, domain, image, and the PEM of the CA that will sign Santa client certificates.
4. `gcloud auth application-default login` (or a service account with permission to create the stack).
5. `ENV=example make tf-init` and `ENV=example make plan`, then apply.

The apply prints `sync_base_url`. Put that value in Santa's `SyncBaseURL`. Certificate Manager also prints a DNS challenge record; the managed server certificate stays provisioning until that record exists. See [docs/deploy.md](docs/deploy.md).

Client certificate issuance, MDM delivery, and the trust config are in [docs/certificates.md](docs/certificates.md).

## CLI

The CLI uses Application Default Credentials and the same Firestore database as the API.

```bash
export ENV=example
./build/rudolph rules --global
./build/rudolph config get --global
```

`ENV` selects `deployments/environments/$ENV/config.auto.tfvars.json`, or `$HOME/.rudolph-cli/$ENV.json` when you are outside the repo. The file must contain `project_id` and `firestore_database`.

## Protocol docs

- [Sync protocol](docs/sync.md)
- [Rules](docs/rules.md)
- [Configuring Santa](docs/configuring-santa.md)
- [Migration from the AWS upstream](docs/migration.md)

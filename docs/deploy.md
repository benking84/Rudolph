# Deployment

Rudolph's Google Cloud stack is the Terraform module in `deployments/terraform_modules/santa_api`. `deployments/environments/example` is a root module that calls it.

## What you need

- Go 1.22+ to build the binary, and Docker if you build the image locally.
- Terraform 1.5+.
- A GCP project where you can enable APIs, create a Cloud Run service, a Firestore Native database, a load balancer, Certificate Manager resources, and IAM bindings.
- A DNS name you control, for example `sync.example.com`.
- The CA certificate (public PEM) that will sign Santa client certificates. Creating that CA and the client certificates is covered in [certificates.md](certificates.md).

The example root module uses a local Terraform backend so `terraform init` does not need a bucket. For a real environment, create a Cloud Storage bucket and switch `versions.tf` to a `gcs` backend before the first apply. Do not commit the state file.

## First environment

```bash
export ENV=prod
make init-env
```

That copies `deployments/environments/example` to `deployments/environments/prod`. Edit `config.auto.tfvars.json`:

| Field | Meaning |
| --- | --- |
| `project_id` | GCP project |
| `region` | Cloud Run, KMS, and Artifact Registry region, such as `us-central1` |
| `firestore_location` | Firestore location. Use the same region, or a multi-region such as `nam5` |
| `prefix` | Short name prefix, 1–19 characters, lowercase |
| `domain_name` | Public hostname, without a scheme |
| `container_image` | Image URI. You can put a placeholder here for the first plan, then replace it after the push below |
| `client_ca_pem` | PEM of the client CA, including the BEGIN/END lines |
| `dns_managed_zone` | Cloud DNS zone *name* if Terraform should create records. Empty means you create them |
| `allowed_inbound_cidrs` | Optional Cloud Armor allow list. Empty allows every source that presents a valid client certificate |
| `enable_event_export` | `true` publishes execution events and archives them |
| `deletion_protection` / `firestore_delete_protection` | Turn these on after the first successful apply |

Authenticate:

```bash
gcloud auth application-default login
gcloud config set project PROJECT_ID
```

Then:

```bash
ENV=prod make tf-init
ENV=prod make plan
```

Review the plan and apply with Terraform in that directory (`terraform apply`). `make deploy` runs tests, builds local binaries, and applies. It does not push the container. Push the image first, or Cloud Run will fail to start.

## Image

Terraform creates an Artifact Registry Docker repository. After apply, its URL is the `artifact_registry` output.

```bash
gcloud auth configure-docker REGION-docker.pkg.dev
docker build -t REGION-docker.pkg.dev/PROJECT/PREFIX-rudolph/rudolph-api:latest .
docker push REGION-docker.pkg.dev/PROJECT/PREFIX-rudolph/rudolph-api:latest
```

Set `container_image` to that URI and apply again if the first apply used a placeholder. The Cloud Run service account does not need to pull the image itself; the Cloud Run service agent does, and it can pull from Artifact Registry in the same project.

## DNS and the server certificate

Apply prints:

- `load_balancer_ip` — A record for `domain_name`.
- `certificate_dns_record_name`, `certificate_dns_record_type`, and `certificate_dns_record_data` — the Certificate Manager challenge.

If `dns_managed_zone` is set, Terraform creates both records in that zone. Otherwise create them at your DNS host. The Google-managed server certificate stays provisioning until the challenge record is visible. Santa should not be pointed at the hostname until the certificate is active.

`sync_base_url` is `https://DOMAIN/`. That string, including the trailing slash, is Santa's `SyncBaseURL`.

## What the apply creates

- Enabled APIs (Cloud Run, Firestore, Compute, Certificate Manager, Network Security, Secret Manager, Pub/Sub, Storage, Artifact Registry, Cloud KMS, IAM, DNS, Logging, Monitoring).
- A Firestore Native database, TTL on `expireAt`, and the indexes the sync and CLI queries use.
- A Cloud Run service with ingress restricted to the load balancer.
- A serverless NEG, global external Application Load Balancer, HTTP-to-HTTPS redirect, managed server certificate, and mTLS trust config.
- Secret Manager secret containing the client CA.
- A Cloud KMS key, a Pub/Sub topic, and a Cloud Storage archive subscription when event export is enabled.
- The runtime service account and its roles.

Cloud Run logs appear in Cloud Logging under the service name. There is no separate log group to create.

## CLI against the same database

`config.auto.tfvars.json` is also what the CLI reads. From the repo:

```bash
export ENV=prod
./build/rudolph config get --global
```

Your user or service account needs `roles/datastore.user` on the project (or a narrower Firestore role) to edit rules. Application Default Credentials are used. No AWS profile remains.

## Destroy

Firestore delete protection and Cloud Run deletion protection block destroy while they are enabled. Set both to false, apply, then destroy. The events bucket is not force-destroyed unless `force_destroy_events_bucket` is true. The Firestore database uses Terraform deletion policy `ABANDON`, so destroying the stack leaves the database in place rather than deleting rule data. Remove it in the console only after you have exported what you need.

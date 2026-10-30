# Example environment

This root module deploys Rudolph on Google Cloud by calling `deployments/terraform_modules/santa_api`.

`config.auto.tfvars.json` is loaded automatically. Replace every example value before you apply:

- `project_id` must be your project.
- `client_ca_pem` in the sample file is a throwaway certificate generated for the example. It is not a CA you should ship. Replace it with the CA that signs your Santa client certificates. The matching private key is not in this repo.
- `container_image` must be an image you have pushed to Artifact Registry. The sample URI will not run.
- `deletion_protection` and `firestore_delete_protection` are false here so a demo stack can be destroyed. Turn them on for a real deployment.
- `dns_managed_zone` is empty. Create the A record and the Certificate Manager DNS challenge yourself, or set the zone name and apply again.

State is local (`terraform.tfstate`). Switch `versions.tf` to a GCS backend before sharing this environment.

```bash
export ENV=example
make tf-init
make plan
```

The CLI reads this same JSON. It accepts `project_id` and either `firestore_database` or `firestore_database_id`.

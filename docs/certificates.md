# Client certificates

Santa authenticates to Rudolph with mutual TLS. The external HTTPS load balancer rejects any connection whose client certificate does not chain to the CA in the Certificate Manager trust config. Cloud Run never sees a failed handshake.

You bring your own CA. Rudolph does not operate a certificate authority and does not enroll devices. The usual production source of the client certificate is the same MDM that installs Santa.

## 1. Create a CA

Use a CA you already trust for device identity, or create one that exists only for Santa sync. The example below is an offline root you can replace with a corporate CA. Keep the private key offline. Terraform needs the certificate, not the key.

```bash
openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 -nodes \
  -keyout rudolph-client-ca.key \
  -out rudolph-client-ca.crt \
  -subj "/CN=Rudolph Client CA"
```

Put the PEM from `rudolph-client-ca.crt` in `client_ca_pem`. If clients present a chain with intermediates, also set `client_intermediate_pems` to those PEM strings, in order. Terraform writes the root PEM to Secret Manager as well as to the trust config.

The load balancer policy is `REJECT_INVALID`. Expired, untrusted, or missing client certificates fail at the proxy. There is no application fallback.

## 2. Issue a certificate per machine

Each Mac should have its own client certificate. A shared certificate works technically and is weaker: you cannot revoke one machine, and the load balancer logs cannot tell machines apart.

A reasonable profile:

- Key usage: digital signature
- Extended key usage: client authentication (`1.3.6.1.5.5.7.3.2`)
- Subject CN or a SAN that identifies the device (serial number, asset tag, or the Santa machine UUID)
- Lifetime aligned with your MDM renewal window

Rudolph does not read the certificate subject to decide which rules to serve. The machine id in the URL is still the Santa `MachineID`. The certificate proves that the caller is a device your CA enrolled. Binding the certificate subject to the machine id would be an extra check you can add later; it is not required for the sync protocol.

Example issuance with the CA from step 1:

```bash
openssl req -newkey rsa:2048 -nodes \
  -keyout machine.key \
  -out machine.csr \
  -subj "/CN=AAAAAAAA-A00A-1234-1234-5864377B4831"

cat > machine.ext <<'EOF'
extendedKeyUsage = clientAuth
keyUsage = digitalSignature
EOF

openssl x509 -req -in machine.csr -CA rudolph-client-ca.crt -CAkey rudolph-client-ca.key \
  -CAcreateserial -out machine.crt -days 825 -sha256 -extfile machine.ext
```

Package `machine.key` and `machine.crt` as a PKCS#12 file if your MDM imports identities that way:

```bash
openssl pkcs12 -export -inkey machine.key -in machine.crt \
  -certfile rudolph-client-ca.crt -out machine.p12 -passout pass:
```

## 3. Deliver the identity with MDM

Install the identity as a device certificate, not a user certificate, so Santa can use it before anyone logs in.

For Jamf, Kandji, Intune, or another Apple MDM, the payload is typically:

1. A PKCS#12 identity payload (or an ACME/SCEP payload if the MDM issues the certificate itself) with the client key and certificate.
2. A trust payload for the Rudolph client CA, only if the device must trust that CA for other reasons. The load balancer does not need the Mac to trust the client CA. The Mac must trust the *server* certificate, which is a public Google-managed certificate for `domain_name` and is already in the system roots once it is active.
3. The Santa configuration profile (`SyncBaseURL` = the `sync_base_url` output).

Santa's sync client uses the identity configured for the sync server. Set the Santa configuration keys your Santa version documents for the sync client certificate (the payload identifier of the identity you installed). Confirm the exact preference key against the Santa version you deploy; the names have changed across releases. The configuration profile in `configs/santa-configuration.mobileconfig` is the place to add that key next to `SyncBaseURL`.

A practical MDM sequence:

1. Deploy the identity profile to a pilot Mac and confirm the certificate lands in the System keychain with a private key.
2. Deploy the Santa configuration profile pointing `SyncBaseURL` at `https://sync.example.com/`.
3. On the pilot Mac, run a sync (`santactl sync`) and confirm Cloud Logging shows `POST /preflight/{machine_id}` with status 200. A TLS failure or HTTP 401 means the certificate did not chain to the trust config, or `REQUIRE_CLIENT_CERT` did not see `X-Client-Cert-Chain-Verified: true`.
4. Roll the identity and the Santa profile out to the fleet. Renew the identity before expiry. Replacing the CA means updating `client_ca_pem` and applying Terraform, then re-issuing client certificates.

## 4. Rotate or revoke

- Revoke one machine by removing its identity in MDM and, if you keep a denylist, blocking that serial at a subscriber or in Cloud Armor. The load balancer trusts every currently valid certificate from the CA; it does not consult a CRL unless you operate that check in front of Rudolph.
- Rotate the CA by adding the new CA to the trust config (apply), issuing new client certificates, and only then removing the old PEM.
- The server certificate rotates by itself. Certificate Manager renews the Google-managed certificate as long as the DNS authorization record stays in place.

## Local calls without a device

`curl` to the public hostname must present a client certificate:

```bash
curl --cert machine.crt --key machine.key \
  "https://sync.example.com/health"
```

`go run ./cmd/api` on a workstation does not require a client certificate unless you set `REQUIRE_CLIENT_CERT=true`. That process is not reachable from the internet. The Cloud Run service sets the variable so anything that slips past a misconfigured proxy is still rejected.


[![Artifact HUB](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/deyaeddin)](https://artifacthub.io/packages/search?repo=deyaeddin)

**Note**: this is based on the project [vadimkim/cert-manager-webhook-hetzner](https://github.com/vadimkim/cert-manager-webhook-hetzner)

# ACME webhook for Hetzner DNS API

This solver can be used when you want to use cert-manager with Hetzner DNS API. API documentation is [here](https://dns.hetzner.com/api-docs)

## Requirements
-   [go](https://golang.org/) >= 1.16.0
-   [helm](https://helm.sh/) >= v3.0.0
-   [kubernetes](https://kubernetes.io/) >= v1.14.0
-   [cert-manager](https://cert-manager.io/) >= v1.3.1

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Webhook

#### Using public helm chart
```bash
helm repo add deyaeddin https://raw.githubusercontent.com/deyaeddin/cert-manager-webhook-hetzner/helmrepo/

"deyaeddin" has been added to your repositories

# now let's install our Chart from our repository
helm install my-cert-manager-webhook-hetzner deyaeddin/cert-manager-webhook-hetzner --version 0.1.x

```

#### From local checkout

```bash
helm install --namespace cert-manager cert-manager-webhook-hetzner chart/cert-manager-webhook-hetzner
```
**Note**: The kubernetes resources used to install the Webhook should be deployed within the same namespace as the cert-manager.

To uninstall the webhook run
```bash
helm uninstall --namespace cert-manager cert-manager-webhook-hetzner
```

## Issuer

Create a `ClusterIssuer` or `Issuer` resource as following:
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: mail@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
      - dns01:
          webhook:
            groupName: acme.yourdomain.here
            solverName: hetzner
            config:
              secretName: hetzner-secret
              zoneName: example.com.
              apiUrl: https://dns.hetzner.com/api/v1
```

### Credentials
In order to access the Hetzner API, the webhook needs an API token.

If you choose another name for the secret than `hetzner-secret`, ensure you modify the value of `secretName` in the `[Cluster]Issuer`.

The secret for the example above will look like this (for encoding your api token into base64, refer to Running the test suite example) :
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hetzner-secret
type: Opaque
data:
  api-key: your-key-base64-encoded
```

**Note: if you are using terraform to create a resource "kubernetes_secret" ... then you should not hash the your-key. 

### Create a certificate

Finally you can create certificates, for example:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-cert
  namespace: cert-manager
spec:
  commonName: example.com
  dnsNames:
    - example.com
  issuerRef:
    name: letsencrypt-staging
    kind: ClusterIssuer
  secretName: example-cert
```

## Development

### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating a
DNS01 webhook.**

First, you need to have Hetzner account with access to DNS control panel. You need to create API token and have a registered and verified DNS zone there.
Then you need to create 2 environment variables:

 - `TEST_ZONE_NAME` to be used as `zoneName` parameter for the generated `testdata/hetzner/config.json` file during the testing.
 - `HCLOUD_DNS_API_TOKEN` to fill out the api-key field in the generated secret `testdata/hetzner/hetzner-secret.yml` file. You must encode your api token into base64 and use the hash 

Example generating encoded api-key hash:
```bash
echo -n xxxxxxxxxxxxxxxxxxxxxxxxxxx | base64
```

You can then run the test suite with:

```bash
export TEST_ZONE_NAME=example.com.
export HCLOUD_DNS_API_TOKEN={result of echo -n xxxxxxxxxxxxxxxxxxxxxxxxxxx | base64}
make test
```

**Note** : resolved FQDN must end with '.', therefore, zoneName must end with the same.

* **If you are forking this, you need to put these variables in repo secrets as used in [testing.yml](https://github.com/deyaeddin/cert-manager-webhook-hetzner/blob/6b1264fc49adad427901a8177f26789be626d352/.github/workflows/testing.yml#L18)** 
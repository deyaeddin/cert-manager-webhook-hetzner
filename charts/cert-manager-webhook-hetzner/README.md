# ACME webhook for Hetzner DNS API

[cert-manager-webhook-hetzner](https://github.com/deyaeddin/cert-manager-webhook-hetzner) is a solver can be used when you want to use cert-manager with Hetzner DNS API. API documentation is [here](https://dns.hetzner.com/api-docs)

## Prerequisites

- Kubernetes 1.12+
- Helm 3.1.0

## Installing the Chart

Add the chart repo to Helm:
```bash
helm repo add deyaeddin https://deyaeddin.github.io/cert-manager-webhook-hetzner/chart/
helm install my-release deyaeddin/cert-manager-webhook-hetzner

```
The installation command will deploy the webhook on the Kubernetes cluster in the default configuration. Please refer to parameters section to adjust.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

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
            groupName: acme.unique.company.name
            solverName: hetzner
            config:
              secretName: hetzner-secret
              zoneName: example.com. # REPLACE THIS WITH YOUR ZONE!!!
              apiUrl: https://dns.hetzner.com/api/v1
```

### Credentials
In order to access the Hetzner API, the webhook needs an API token.

If you choose another name for the secret than `hetzner-secret`, ensure you modify the value of `secretName` in the `[Cluster]Issuer`.

The secret for the example above will look like this :
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hetzner-secret
type: Opaque
data:
  api-key: your-key-base64-encoded
```

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


## Parameters

The following table lists the configurable parameters of the cert-manager-webhook-hetzner chart, and their default values.

| Parameter                          | Description                                     | Default                                                 |
|------------------------------------|-------------------------------------------------|---------------------------------------------------------|
| `groupName`                        | Group name for the webhook                      | `acme.unique.company.name`                              |
| `certManager.namespace`            | cert-manager namespace                          | `cert-manager`                                          |
| `certManager.serviceAccountName`   | cert-manager service account name               | `cert-manager`                                          |
| `image.repository`                 | Docker image repository                         | `deyaeddin/cert-manager-webhook-hetzner`                |
| `image.tag`                        | Docker image tag                                | `latest`                                                |
| `image.pullPolicy`                 | Docker image pull policy                        | `IfNotPresent`                                          |
| `replicaCount`                     | Number of webhook replicas to deploy            | `1`                                                     |
| `nameOverride`                     | Name override for the chart                     | `""`                                                    |
| `fullnameOverride`                 | Full name override for the chart                | `""`                                                    |
| `service.type`                     | Service type                                    | `ClusterIP`                                             |
| `service.port`                     | Service port                                    | `443`                                                   |
| `secretName`                       | secret name created in Credentials              | `hetzner-secret`                                        |
| `resources`                        | Pod resources                                   | Check `values.yaml` file                                |
| `nodeSelector`                     | Node selector                                   | `nil`                                                   |
| `tolerations`                      | Node toleration                                 | `nil`                                                   |
| `affinity`                         | Node affinity                                   | `nil`                                                   |
| `podSecurityContext`               | webhook pods' Security Context                  | Check `values.yaml` file                                |
| `containerSecurityContext`         | webhook containers' Security Context            | Check `values.yaml` file                                |


**Useful links**

- https://cert-manager.io/docs/configuration/acme/dns01/#webhook
- https://helm.sh/docs/chart_best_practices/values/
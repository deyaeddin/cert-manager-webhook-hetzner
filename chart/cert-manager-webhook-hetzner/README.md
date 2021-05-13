# ACME webhook for Hetzner DNS API

[cert-manager-webhook-hetzner](https://github.com/prometheus/node_exporter) is a solver can be used when you want to use cert-manager with Hetzner DNS API. API documentation is [here](https://dns.hetzner.com/api-docs)

## TL;DR

```bash
$ helm repo add deyaeddin https://github.com/deyaeddin/cert-manager-webhook-hetzner
$ helm install my-release deyaeddin/cert-manager-webhook-hetzner
```

## Introduction

This chart bootstraps [cert-manager-webhook-hetzner](https://github.com/deyaeddin/cert-manager-webhook-hetzner) on [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Kubernetes 1.12+
- Helm 3.1.0

## Installing the Chart

Add the chart repo to Helm:
```bash
helm repo add cert-manager-webhook-hetzner https://raw.githubusercontent.com/deyaeddin/cert-manager-webhook-hetzner/helmrepo/

"cert-manager-webhook-hetzner" has been added to your repositories

# now let's install our Chart from our repository
helm install -n cert-manager cert-manager-webhook-hetzner deyaeddin/cert-manager-webhook-hetzner

```
#TODO:
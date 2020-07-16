[![goreleaser](https://github.com/itzg/kube-metrics-exporter/workflows/goreleaser/badge.svg)](https://github.com/itzg/kube-metrics-exporter/actions?query=workflow%3Agoreleaser)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/itzg/kube-metrics-exporter)](https://github.com/itzg/kube-metrics-exporter/releases/latest)
[![Docker Pulls](https://img.shields.io/docker/pulls/itzg/kube-metrics-exporter)](https://hub.docker.com/r/itzg/kube-metrics-exporter)

Simple application that accesses the [Kubernetes metrics API](https://github.com/kubernetes/metrics) and exports the pod metrics for Prometheus scraping.

The Metrics API is exposed by a deployed [Metrics Server](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-metrics-pipeline/#metrics-server) which is included in most managed clusters. [It can also be deployed separately.](https://github.com/kubernetes-sigs/metrics-server).

## Metrics

This services exports two metrics:
- `container_cpu_usage`
- `container_mem_usage`

Both metrics include the labels:
- `namespace`
- `pod`
- `container`

### Prometheus label renaming

By default, Prometheus will rename the labels above to avoid conflicts with the same labels applied during export. As a result, the metric in Prometheus will appear as:

```
container_cpu_usage{container="kube-metrics-exporter",endpoint="http",exported_container="grafana",exported_namespace="default",exported_pod="grafana-0",instance="10.40.1.109:8080",job="thanos-poc/monitor-metrics-http",namespace="default",pod="kube-metrics-exporter-6d9b8f978d-84x6q"}
```

### Example
```
# HELP container_cpu_usage millicores of CPU used
# TYPE container_cpu_usage gauge
container_cpu_usage{container="grafana",namespace="default",pod="grafana-0"} 2
# HELP container_mem_usage mebibytes of memory used
# TYPE container_mem_usage gauge
container_mem_usage{container="grafana",namespace="default",pod="grafana-0"} 25
```

## Command-line

```
  -debug
        enable debug logging (env DEBUG)
  -http-binding string
        binding of http listener for metrics export (env HTTP_BINDING) (default
":8080")
  -metrics-path string
        http path for metrics export (env METRICS_PATH) (default "/metrics")
  -namespace string
        the namespace of the pods to collect (env NAMESPACE) (default "default")
```

## Stand-alone Usage

The `kube-metrics-exporter` executable can be executed outside of Kubernetes cluster, in which case it will locate and use the kubernetes configuration from the standard location(s).

## In-cluster Usage

With a service account defined with the correct roles, [as described below](#service-account), the reporter can be deployed with a pod manifest such as the following to export metrics for pods in the same namespace:

```yaml
    metadata:
      name: kube-metrics-exporter
      labels:
        app: kube-metrics-exporter
    spec:
      serviceAccountName: kube-metrics-monitor
      containers:
        - name: kube-metrics-exporter
          image: itzg/kube-metrics-exporter
          env:
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

## Service account

Since this application accesses the metrics API of the kubernetes API service, the pod will need to be assigned a service account with an appropriate role. 

> Service accounts must be present before the deployment, so either ensure the service account manifest is applied first or place the service account yaml documents before the deployment in the same manifest file.

The following shows how a service account could be declared:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-metrics-monitor
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kube-metrics-monitor
rules:
  - apiGroups: ["metrics.k8s.io"]
    resources:
      - pods
    verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: kube-metrics-monitor
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: kube-metrics-monitor
subjects:
  - kind: ServiceAccount
    name: kube-metrics-monitor
```

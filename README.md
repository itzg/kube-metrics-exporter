Simple application that accesses the [Kubernetes metrics API](https://github.com/kubernetes/metrics) and exports them for Prometheus scraping.

The Metrics API is exposed by a deployed [Metrics Server](https://kubernetes.io/docs/tasks/debug-application-cluster/resource-metrics-pipeline/#metrics-server) which is included in most managed clusters. [It can also be deployed separately.](https://github.com/kubernetes-sigs/metrics-server).

## Stand-alone Usage

The `kube-metrics-exporter` executable can be executed outside of Kubernetes cluster, in which case it will locate and use the kubernetes configuration from the standard location(s).

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

## In-cluster Usage

With a service account defined with the correct roles, [as described below](#service-account), the reporter can be deployed with a pod manifest such as the following:

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

## Releasing

Create a semver tag with a "v" prefix and push, such as:

```shell script
git tag -s v1.0.0 -m v1.0.0
git push origin v1.0.0
```

A [GitHub Action](https://github.com/itzg/kube-metrics-exporter/actions/runs/171729985) will take of running goreleaser, which will [create a release](https://github.com/itzg/kube-metrics-exporter/releases) and [push an image](https://hub.docker.com/r/itzg/kube-metrics-exporter).
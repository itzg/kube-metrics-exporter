module github.com/itzg/kube-metrics-exporter

go 1.14

require (
	github.com/itzg/go-flagsfiller v1.4.1
	github.com/itzg/zapconfigs v0.1.0
	github.com/prometheus/client_golang v1.7.1
	go.uber.org/zap v1.13.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/metrics v0.17.0
)

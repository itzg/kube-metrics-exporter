package pkg

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type KubeMetricsCollector struct {
	logger     *zap.Logger
	clientset  *versioned.Clientset
	namespaces []string
}

func NewKubeMetricsCollector(logger *zap.Logger, clientset *versioned.Clientset, namespaces []string) *KubeMetricsCollector {
	return &KubeMetricsCollector{
		logger:     logger,
		clientset:  clientset,
		namespaces: namespaces,
	}
}

const (
	LabelNamespace = "namespace"
	LabelPod       = "pod"
	LabelContainer = "container"
)

var (
	CommonLabels = []string{LabelNamespace, LabelPod, LabelContainer}
	DescCpuUsage = prometheus.NewDesc("container_cpu_usage_cores", "CPU cores used", CommonLabels, nil)
	DescMemUsage = prometheus.NewDesc("container_memory_usage_bytes", "memory used", CommonLabels, nil)
)

func (c *KubeMetricsCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- DescCpuUsage
	descs <- DescMemUsage
}

func (c *KubeMetricsCollector) UpdateNamespaces(namespaces []string) {
	c.namespaces = namespaces
}

func (c *KubeMetricsCollector) Collect(metrics chan<- prometheus.Metric) {
	for _, namespace := range c.namespaces {
		podMetricsAccessor := c.clientset.MetricsV1beta1().PodMetricses(namespace)

		podMetricsList, err := podMetricsAccessor.List(v1.ListOptions{})
		if err != nil {
			c.logger.Warn(
				"failed to list kube-metrics",
				zap.String("namespace", namespace),
				zap.Error(err))
			continue
		}

		for _, p := range podMetricsList.Items {
			podName := p.Name
			for _, container := range p.Containers {
				containerName := container.Name
				// matching the units reported by kubectl top pods
				cpuUsage := container.Usage.Cpu().ScaledValue(resource.Milli)
				memUsage := container.Usage.Memory().Value()

				metric, err := prometheus.NewConstMetric(
					DescCpuUsage,
					prometheus.GaugeValue,
					float64(cpuUsage)/1000,
					// labels
					namespace,
					podName,
					containerName)
				if err != nil {
					c.logger.Warn("failed to create metric", zap.Error(err))
				} else {
					metrics <- metric
				}

				metric, err = prometheus.NewConstMetric(
					DescMemUsage,
					prometheus.GaugeValue,
					float64(memUsage),
					// labels
					namespace,
					podName,
					containerName)
				if err != nil {
					c.logger.Warn("failed to create metric", zap.Error(err))
				} else {
					metrics <- metric
				}
			}
		}
	}
}

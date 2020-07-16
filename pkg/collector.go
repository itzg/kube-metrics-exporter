package pkg

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type KubeMetricsCollector struct {
	podMetricsAccessor v1beta1.PodMetricsInterface
	logger             *zap.Logger
	namespace          string
}

func NewKubeMetricsCollector(podMetricsAccessor v1beta1.PodMetricsInterface, logger *zap.Logger, namespace string) *KubeMetricsCollector {
	return &KubeMetricsCollector{
		podMetricsAccessor: podMetricsAccessor,
		logger:             logger,
		namespace:          namespace,
	}
}

const (
	LabelNamespace = "namespace"
	LabelPod       = "pod"
	LabelContainer = "container"
)

var (
	CommonLabels = []string{LabelNamespace, LabelPod, LabelContainer}
	DescCpuUsage = prometheus.NewDesc("pod_cpu_usage", "millicores of CPU used", CommonLabels, nil)
	DescMemUsage = prometheus.NewDesc("pod_mem_usage", "mebibytes of memory used", CommonLabels, nil)
)

func (c *KubeMetricsCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- DescCpuUsage
	descs <- DescMemUsage
}

func (c *KubeMetricsCollector) Collect(metrics chan<- prometheus.Metric) {
	podMetricsList, err := c.podMetricsAccessor.List(v1.ListOptions{})
	if err != nil {
		c.logger.Error("failed to list kube metrics", zap.Error(err))
		return
	}

	for _, p := range podMetricsList.Items {
		podName := p.Name
		for _, container := range p.Containers {
			containerName := container.Name
			// matching the units reported by kubectl top pods
			cpuUsage := container.Usage.Cpu().ScaledValue(resource.Milli)
			memUsage := container.Usage.Memory().ScaledValue(resource.Mega)

			metric, err := prometheus.NewConstMetric(DescCpuUsage, prometheus.GaugeValue, float64(cpuUsage),
				c.namespace, podName, containerName)
			if err != nil {
				c.logger.Warn("failed to create metric", zap.Error(err))
			} else {
				metrics <- metric
			}

			metric, err = prometheus.NewConstMetric(DescMemUsage, prometheus.GaugeValue, float64(memUsage),
				c.namespace, podName, containerName)
			if err != nil {
				c.logger.Warn("failed to create metric", zap.Error(err))
			} else {
				metrics <- metric
			}
		}
	}

}

package pkg

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	rest "k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// configuration for the exporter
type Config struct {
	Namespace        string   `default:"default" usage:"the namespace of the pods to collect"`
	IgnoreNamespaces []string `default:"kube-system" usage:"when 'namespace' is empty, this lists namespaces to ignore"`
	HttpBinding      string   `usage:"binding of http listener for metrics export" default:":8080"`
	MetricsPath      string   `usage:"http path for metrics export" default:"/metrics"`
	Debug            bool     `usage:"enable debug logging"`
}

type Handler struct {
	ExporterConfig Config
	Logger         *zap.Logger
	ClientSet      *versioned.Clientset
	KubeConfig     *rest.Config
	collector      *KubeMetricsCollector
}

func (h *Handler) SetupCollector(mu *sync.RWMutex) (err error) {
	h.collector = NewKubeMetricsCollector(h.Logger.Named("collector"), h.ClientSet, mu)
	err = prometheus.Register(h.collector)
	if err != nil {
		return
	}

	h.Logger.Info("registered collector")
	return
}

func (h *Handler) UpdateNamespaces(namespaces []string) {
	h.collector.UpdateNamespaces(namespaces)
}

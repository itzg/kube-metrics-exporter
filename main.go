package main

import (
	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/kube-metrics-exporter/pkg"
	"github.com/itzg/zapconfigs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"net/http"
)

var config struct {
	Namespace   string `default:"default" usage:"the namespace of the pods to collect"`
	HttpBinding string `usage:"binding of http listener for metrics export" default:":8080"`
	MetricsPath string `usage:"http path for metrics export" default:"/metrics"`
	Debug       bool   `usage:"enable debug logging"`
}

func main() {

	err := flagsfiller.Parse(&config, flagsfiller.WithEnv(""))
	if err != nil {
		log.Fatal(err)
	}

	var logger *zap.Logger
	if config.Debug {
		logger = zapconfigs.NewDebugLogger()
	} else {
		logger = zapconfigs.NewDefaultLogger()
	}
	defer logger.Sync()

	// Connect to kubernetes and get metrics clientset

	configLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
	if err != nil {
		logger.Fatal("loading kubeConfig", zap.Error(err))
	}

	clientset, err := versioned.NewForConfig(kubeConfig)
	if err != nil {
		logger.Fatal("creating metrics clientset", zap.Error(err))
	}

	podMetricsAccessor := clientset.MetricsV1beta1().PodMetricses(config.Namespace)

	err = prometheus.Register(pkg.NewKubeMetricsCollector(podMetricsAccessor, logger.Named("collector"), config.Namespace))
	if err != nil {
		logger.Fatal("collector registration failed", zap.Error(err))
	}

	http.Handle(config.MetricsPath, promhttp.Handler())

	logger.Info("ready to export metrics", zap.String("binding", config.HttpBinding),
		zap.String("path", config.MetricsPath),
		zap.String("namespace", config.Namespace),
	)
	err = http.ListenAndServe(config.HttpBinding, nil)
	logger.Fatal("http server failed", zap.Error(err))
}

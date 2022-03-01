package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/itzg/go-flagsfiller"
	"github.com/itzg/kube-metrics-exporter/pkg"
	"github.com/itzg/zapconfigs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

var mu sync.RWMutex
var namespaces []string

var config pkg.Config

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

	handler := pkg.Handler{
		Logger:         logger,
		ExporterConfig: config,
		KubeConfig:     kubeConfig,
		ClientSet:      clientset,
	}

	namespaces, err = handler.GetNamespaces()
	if err != nil {
		logger.Fatal("failed getting namespaces", zap.Error(err))
	}

	err = handler.SetupCollector(namespaces)
	if err != nil {
		logger.Fatal("could not register collectors", zap.Error(err))
	}

	go func() {
		for {
			logger.Info("checking namespaces...")
			time.Sleep(time.Second * 20)
			mu.Lock()
			namespaces, err = handler.GetNamespaces()
			if err != nil {
				logger.Error("failed getting namespaces", zap.Error(err))
			}

			logger.Info("Updating collector with new namespaces...")
			handler.UpdateNamespaces(namespaces)
			mu.Unlock()
		}
	}()

	logger.Info(
		fmt.Sprintf("Starting http://0.0.0.0%s/%s", config.HttpBinding, config.MetricsPath))
	http.Handle(config.MetricsPath, promhttp.Handler())

	err = http.ListenAndServe(config.HttpBinding, nil)
	logger.Fatal("http server failed", zap.Error(err))
}

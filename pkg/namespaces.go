package pkg

import (
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (h *Handler) GetNamespaces() (ns []string, err error) {
	if h.ExporterConfig.Namespace != "" {
		ns = append(ns, h.ExporterConfig.Namespace)
		return
	}

	k8sClient, err := kubernetes.NewForConfig(h.KubeConfig)
	if err != nil {
		return
	}

	namespaces, err := k8sClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return
	}

	h.Logger.Debug("retrieved namespaces from k8s", zap.Int("count", len(namespaces.Items)))

	for _, namespace := range namespaces.Items {
		if contains(h.ExporterConfig.IgnoreNamespaces, namespace.Name) {
			h.Logger.Debug("skipping namespace", zap.String("namespace", namespace.Name))
			continue
		}

		h.Logger.Debug("adding namespace to watch", zap.String("namespace", namespace.Name))
		ns = append(ns, namespace.Name)
	}

	return
}

func contains(sl []string, name string) bool {
	for _, value := range sl {
		if value == name {
			return true
		}
	}
	return false
}

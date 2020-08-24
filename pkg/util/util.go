package util

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetClientSetFromConfig GetClientSetFromConfig
func GetClientSetFromConfig(cfg *rest.Config) (*kubernetes.Clientset, error) {

	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

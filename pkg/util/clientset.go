package util

import (
	"k8s.io/client-go/kubernetes"
)

// MustGetClientSet MustGetClientSet
func MustGetClientSet() *kubernetes.Clientset {
	clientset, err := GetClientSet()
	if err != nil {
		panic(err)
	}

	return clientset
}

// GetClientSet GetClientSet
func GetClientSet() (*kubernetes.Clientset, error) {
	cfg, err := GetKubeconfig()
	if err != nil {
		return nil, err
	}
	setKubernetesDefaults(cfg)
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

package util

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
)

// MustGetKubeconfig MustGetKubeconfig
func MustGetKubeconfig() *rest.Config {
	cfg, err := GetKubeconfig()
	if err != nil {
		panic(err)
	}
	return cfg
}

// GetKubeconfig GetKubeconfig
func GetKubeconfig() (*rest.Config, error) {
	kubeconfigPath, err := GetKubeconfigPath()
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	setKubernetesDefaults(cfg)
	return cfg, nil
}

func setKubernetesDefaults(config *rest.Config) error {
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}

	if config.APIPath == "" {
		config.APIPath = "/api"
	}
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	}
	return rest.SetKubernetesDefaults(config)
}

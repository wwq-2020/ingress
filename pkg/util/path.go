package util

import (
	"os"
	"path"
)

// MustGetKubeconfigPath MustGetKubeconfigPath
func MustGetKubeconfigPath() string {
	kubeconfigPath, err := GetKubeconfigPath()
	if err != nil {
		panic(err)
	}
	return kubeconfigPath
}

// GetKubeconfigPath GetKubeconfigPath
func GetKubeconfigPath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(userHomeDir, ".kube", "config"), nil
}

package kube

import (
	"fmt"
	"github.com/romana/rlog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

const (
	TokenFilePath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

func IsRunningOutOfKubeCluster() bool {
	_, err := os.Stat(TokenFilePath)
	return os.IsNotExist(err)
}

type Kube struct {
	Client kubernetes.Interface
}

// TODO: check reconnection to kubernetes
func NewKube() (*Kube, error) {
	var err error
	var config *rest.Config

	if IsRunningOutOfKubeCluster() {
		var kubeconfig string
		if kubeconfig = os.Getenv("KUBECONFIG"); kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}

		rlog.Infof("Kube: using out-of-cluster kubernetes configuration at %s", kubeconfig)

		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("kubernetes out-of-cluster configuration problem: %s", err)
		}
	} else {
		rlog.Info("Kube: using in-cluster kubernetes configuration")

		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("kubernetes in-cluster configuration problem: %s", err)
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("kubernetes connection problem: %s", err)
	}

	rlog.Info("Kube: successfully configured kubernetes")

	return &Kube{Client: client}, nil
}

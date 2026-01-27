// Package client provides Kubernetes client configuration utilities.
//
// Author: Daksh Pathak
// GitHub: https://github.com/dakshhhhh16
// Date: January 2026

package client

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return ""
}

// GetClientSetWithContext creates a Kubernetes clientset using the specified context.
// If contextName is empty, the current context from kubeconfig is used.
// Falls back to in-cluster config if kubeconfig is not available.
func GetClientSetWithContext(contextName string) (*kubernetes.Clientset, error) {
	var (
		config    *rest.Config
		clientset *kubernetes.Clientset
		err       error
	)

	kubeconfig := os.Getenv("KUBECONFIG")

	if kubeconfig == "" {
		if home := homeDir(); home != "" {
			kubeconfig = fmt.Sprintf("%s/.kube/config", home)
		}
	}

	if _, err := os.Stat(kubeconfig); err == nil {
		rawConfig, err := clientcmd.LoadFromFile(kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
		}

		if contextName == "" {
			contextName = rawConfig.CurrentContext
		}

		ctxContext := rawConfig.Contexts[contextName]
		if ctxContext == nil {
			return nil, fmt.Errorf("context '%s' not found in kubeconfig", contextName)
		}

		clientConfig := clientcmd.NewDefaultClientConfig(
			*rawConfig,
			&clientcmd.ConfigOverrides{
				CurrentContext: contextName,
			},
		)

		config, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create REST config: %w", err)
		}
	} else {
		// Fall back to in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}
// k8s client

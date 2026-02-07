package kube

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes dynamic client, typed clientset, and REST config.
type Client struct {
	dynamic    dynamic.Interface
	clientset  kubernetes.Interface
	restConfig *rest.Config
}

func (c *Client) Dynamic() dynamic.Interface      { return c.dynamic }
func (c *Client) Clientset() kubernetes.Interface   { return c.clientset }
func (c *Client) RestConfig() *rest.Config          { return c.restConfig }

// NewClient creates a Kubernetes client using the default kubeconfig loading
// rules (KUBECONFIG env var or ~/.kube/config).
func NewClient() (*Client, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("creating dynamic client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}

	return &Client{
		dynamic:    dynamicClient,
		clientset:  clientset,
		restConfig: restConfig,
	}, nil
}

// EnsureNamespace creates the namespace if it does not already exist.
func (c *Client) EnsureNamespace(ctx context.Context, name string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("checking namespace %q: %w", name, err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err = c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("creating namespace %q: %w", name, err)
	}
	return nil
}

package kube

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

// WaitForReady watches a resource in the agentsandbox.dev/v1 group until its
// Ready condition becomes True or the context is cancelled/times out.
// The resource parameter is the plural resource name (e.g. "sandboxclaims", "sandboxwarmpools").
func (c *Client) WaitForReady(ctx context.Context, namespace, resource, name string) error {
	gvr := schema.GroupVersionResource{
		Group:    "agentsandbox.dev",
		Version:  "v1",
		Resource: resource,
	}

	watcher, err := c.Dynamic().Resource(gvr).Namespace(namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		return fmt.Errorf("watching %s %s/%s: %w", resource, namespace, name, err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for %s %s/%s to become ready", resource, namespace, name)
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed for %s %s/%s", resource, namespace, name)
			}
			if event.Type == watch.Error {
				return fmt.Errorf("watch error for %s %s/%s", resource, namespace, name)
			}

			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}

			if isReady(obj) {
				return nil
			}
		}
	}
}

// isReady checks whether an unstructured object has a condition with
// type=Ready and status=True.
func isReady(obj *unstructured.Unstructured) bool {
	status, found, err := unstructured.NestedMap(obj.Object, "status")
	if err != nil || !found {
		return false
	}

	conditionsRaw, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil || !found {
		return false
	}

	for _, c := range conditionsRaw {
		condition, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _ := condition["type"].(string)
		condStatus, _ := condition["status"].(string)
		if condType == "Ready" && condStatus == "True" {
			return true
		}
	}

	return false
}

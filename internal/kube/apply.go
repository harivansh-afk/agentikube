package kube

import (
	"bytes"
	"context"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/restmapper"
	yamlserializer "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// ServerSideApply splits a multi-document YAML into individual resources
// and applies each one using server-side apply with the "agentikube" field manager.
func (c *Client) ServerSideApply(ctx context.Context, manifests []byte) error {
	decoder := k8syaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifests), 4096)

	discoveryClient, ok := c.Clientset().Discovery().(*discovery.DiscoveryClient)
	if !ok {
		return fmt.Errorf("failed to get discovery client")
	}
	cachedDiscovery := memory.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscovery)

	deserializer := yamlserializer.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	for {
		var rawObj unstructured.Unstructured
		if err := decoder.Decode(&rawObj); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("decoding YAML document: %w", err)
		}

		// Skip empty documents
		if len(rawObj.Object) == 0 {
			continue
		}

		// Re-encode to JSON for the patch body
		rawJSON, err := rawObj.MarshalJSON()
		if err != nil {
			return fmt.Errorf("marshaling to JSON: %w", err)
		}

		// Decode to get the GVK
		obj := &unstructured.Unstructured{}
		_, gvk, err := deserializer.Decode(rawJSON, nil, obj)
		if err != nil {
			return fmt.Errorf("deserializing object: %w", err)
		}

		// Map GVK to GVR using the REST mapper
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return fmt.Errorf("mapping GVK %s to GVR: %w", gvk.String(), err)
		}

		gvr := mapping.Resource
		name := obj.GetName()
		namespace := obj.GetNamespace()

		applyOpts := metav1.ApplyOptions{
			FieldManager: "agentikube",
		}

		// Apply using the dynamic client - handle namespaced vs cluster-scoped
		if namespace != "" {
			_, err = c.Dynamic().Resource(gvr).Namespace(namespace).Patch(
				ctx, name, types.ApplyPatchType, rawJSON, applyOpts.ToPatchOptions(),
			)
		} else {
			_, err = c.Dynamic().Resource(gvr).Patch(
				ctx, name, types.ApplyPatchType, rawJSON, applyOpts.ToPatchOptions(),
			)
		}
		if err != nil {
			return fmt.Errorf("applying %s/%s: %w", gvk.Kind, name, err)
		}

		fmt.Printf("applied %s/%s\n", gvk.Kind, name)
	}

	return nil
}

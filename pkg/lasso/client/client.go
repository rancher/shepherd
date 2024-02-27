package client

import (
	"context"
	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/shepherd/pkg/session"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

// Client performs CRUD like operations on a specific GVR.
type Client struct {
	*client.Client
	Session *session.Session
}

// Create will attempt to create the provided object in the given namespace (if client.Namespaced is set to true).
// Create will then attempt to unmarshal the created object from the response into the provide result object.
// If the returned response object is of type Status and has .Status != StatusSuccess, the
// additional information in Status will be used to enrich the error.
const (
	ErrNotFound      = "404 Not Found"
	ErrFailedSelfURL = "failed to find self URL of [&{  map[] map[]}]"
)

func (c *Client) Create(ctx context.Context, namespace string, obj, result runtime.Object, opts metav1.CreateOptions) (err error) {
	err = c.Create(ctx, namespace, obj, result, opts)

	// Unstructure result runtime object
	tempObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(result)
	if err != nil {
		return err
	}

	// Create Unstructured object
	unstructuredTempObj := unstructured.Unstructured{Object: tempObj}

	name := unstructuredTempObj.GetName()

	c.Session.RegisterCleanupFunc(func() error {
		return c.handleCleanup(ctx, namespace, name)
	})
	return nil
}

func (c *Client) handleCleanup(ctx context.Context, namespace string, name string) error {
	err := c.Delete(ctx, namespace, name, metav1.DeleteOptions{})
	if err != nil && (strings.Contains(err.Error(), ErrNotFound) || strings.Contains(err.Error(), ErrFailedSelfURL)) {
		return nil
	}
	return err
}

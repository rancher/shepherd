package clients

import (
	"context"
	"fmt"
	"github.com/rancher/shepherd/pkg/session"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"time"

	"github.com/rancher/shepherd/pkg/wrangler"
	"github.com/rancher/wrangler/v2/pkg/ratelimit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	"k8s.io/client-go/tools/clientcmd"
)

type Clients struct {
	*wrangler.Context
	Dynamic dynamic.Interface
	ts      *session.Session

	// Ctx is canceled when the Close() is called
	Ctx     context.Context
	cancel  func()
	onClose []func()
}

func (c *Clients) Close() {
	for i := len(c.onClose); i > 0; i-- {
		c.onClose[i-1]()
	}
	c.cancel()
}

// OnClose appends a function to the slice of cleanup functions that will be executed when Close is called.
// It takes a function f as a parameter.
func (c *Clients) OnClose(f func()) {
	c.onClose = append(c.onClose, f)
}

// ForCluster retrieves the client configuration for a specific cluster.
// It takes the namespace, name, and session parameters.
// It returns a new instance of the Clients struct associated with the specified cluster and an error if any occurred.
func (c *Clients) ForCluster(namespace, name string, ts *session.Session) (*Clients, error) {
	secret, err := c.Core.Secret().Get(namespace, name+"-kubeconfig", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	config, err := clientcmd.NewClientConfigFromBytes(secret.Data["value"])
	if err != nil {
		return nil, err
	}

	return NewForConfig(c.Ctx, config, ts)
}

// NewForConfig returns a new instance of Clients based on the provided information.
//
// It takes the following parameters:
// - ctx: The context.Context used for cancellation.
// - config: The clientcmd.ClientConfig used for retrieving the client configuration.
// - ts: The *session.Session used for cleanup registration.
//
// NewForConfig creates a new context with cancellation, retrieves the client configuration
// from the provided config, sets a timeout of 30 minutes and disables rate limiting on the REST config.
// It then creates a new wrangler context and dynamic client from the retrieved REST config.
// Finally, it returns a new instance of Clients with the created wrangler context, dynamic client,
// the provided context, cancellation function, and *session.Session.
// In case of any errors during the process, it cancels the context and returns the error.
func NewForConfig(ctx context.Context, config clientcmd.ClientConfig, ts *session.Session) (*Clients, error) {
	ctx, cancel := context.WithCancel(ctx)

	rest, err := config.ClientConfig()
	if err != nil {
		cancel()
		return nil, err
	}

	rest.Timeout = 30 * time.Minute
	rest.RateLimiter = ratelimit.None

	wranglerCtx, err := wrangler.NewContext(ctx, config, rest)
	if err != nil {
		cancel()
		return nil, err
	}

	dynamic, err := dynamic.NewForConfig(rest)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Clients{
		Context: wranglerCtx,
		Dynamic: dynamic,
		Ctx:     ctx,
		cancel:  cancel,
		ts:      ts,
	}, nil
}

// ResourceClient has dynamic.ResourceInterface embedded so dynamic.ResourceInterface's Create can be overwritten.
type ResourceClient struct {
	dynamic.ResourceInterface
	ts *session.Session
}

var (
	// some GVKs are special and cannot be cleaned up because they do not exist
	// after being created (eg: SelfSubjectAccessReview). We'll not register
	// cleanup functions when creating objects of these kinds.
	noCleanupGVKs = []schema.GroupVersionKind{
		{
			Group:   "authorization.k8s.io",
			Version: "v1",
			Kind:    "SelfSubjectAccessReview",
		},
	}
)

// needsCleanup checks if the provided `*unstructured.Unstructured` object needs cleanup
func needsCleanup(obj *unstructured.Unstructured) bool {
	for _, gvk := range noCleanupGVKs {
		if obj.GroupVersionKind() == gvk {
			return false
		}
	}
	return true
}

// Create is dynamic.ResourceInterface's Create function, that is being overwritten to register its delete function to the session.Session
// that is being reference.
func (c *ResourceClient) Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	unstructuredObj, err := c.ResourceInterface.Create(ctx, obj, opts, subresources...)
	if err != nil {
		return nil, err
	}

	if needsCleanup(obj) {
		c.ts.RegisterCleanupFunc(func() error {
			err := c.Delete(context.TODO(), unstructuredObj.GetName(), metav1.DeleteOptions{}, subresources...)
			if errors.IsNotFound(err) {
				return nil
			}

			name := unstructuredObj.GetName()
			if unstructuredObj.GetNamespace() != "" {
				name = unstructuredObj.GetNamespace() + "/" + name
			}
			gvk := unstructuredObj.GetObjectKind().GroupVersionKind()

			return fmt.Errorf("unable to delete (%v) %v: %w", gvk, name, err)
		})
	}

	return unstructuredObj, err
}

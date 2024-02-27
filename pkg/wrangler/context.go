/*
Package wrangler contains functions for creating a management context with wrangler controllers.
*/
package wrangler

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	fleetv1alpha1api "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/lasso/pkg/controller"
	"github.com/rancher/lasso/pkg/dynamic"
	"github.com/rancher/norman/types"
	catalogv1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	clusterv3api "github.com/rancher/rancher/pkg/apis/cluster.cattle.io/v3"
	managementv3api "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	projectv3api "github.com/rancher/rancher/pkg/apis/project.cattle.io/v3"
	provisioningv1api "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	rkev1api "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/remotedialer"
	"github.com/rancher/shepherd/pkg/generated/controllers/catalog.cattle.io"
	catalogcontrollers "github.com/rancher/shepherd/pkg/generated/controllers/catalog.cattle.io/v1"
	capi "github.com/rancher/shepherd/pkg/generated/controllers/cluster.x-k8s.io"
	"github.com/rancher/shepherd/pkg/generated/controllers/cluster.x-k8s.io/v1beta1"
	"github.com/rancher/shepherd/pkg/generated/controllers/fleet.cattle.io"
	fleetv1alpha1 "github.com/rancher/shepherd/pkg/generated/controllers/fleet.cattle.io/v1alpha1"
	"github.com/rancher/shepherd/pkg/generated/controllers/management.cattle.io"
	managementv3 "github.com/rancher/shepherd/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/rancher/shepherd/pkg/generated/controllers/provisioning.cattle.io"
	provisioningv1 "github.com/rancher/shepherd/pkg/generated/controllers/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/pkg/generated/controllers/rke.cattle.io"
	rkecontrollers "github.com/rancher/shepherd/pkg/generated/controllers/rke.cattle.io/v1"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/wrangler/v2/pkg/apply"
	admissionreg "github.com/rancher/wrangler/v2/pkg/generated/controllers/admissionregistration.k8s.io"
	admissionregcontrollers "github.com/rancher/wrangler/v2/pkg/generated/controllers/admissionregistration.k8s.io/v1"
	crdv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/apiextensions.k8s.io/v1"
	apiregv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/apiregistration.k8s.io/v1"
	"github.com/rancher/wrangler/v2/pkg/generated/controllers/apps"
	appsv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/apps/v1"
	"github.com/rancher/wrangler/v2/pkg/generated/controllers/batch"
	batchv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/batch/v1"
	"github.com/rancher/wrangler/v2/pkg/generated/controllers/core"
	corev1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/v2/pkg/generated/controllers/rbac"
	rbacv1 "github.com/rancher/wrangler/v2/pkg/generated/controllers/rbac/v1"
	"github.com/rancher/wrangler/v2/pkg/generic"
	"github.com/rancher/wrangler/v2/pkg/leader"
	"github.com/rancher/wrangler/v2/pkg/schemes"
	"github.com/sirupsen/logrus"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	apiregistrationv12 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	capiv1beta1api "sigs.k8s.io/cluster-api/api/v1beta1"
)

type controllerContextType string

const (
	Management controllerContextType = "mgmt"
)

var (
	localSchemeBuilder = runtime.SchemeBuilder{
		provisioningv1api.AddToScheme,
		capiv1beta1api.AddToScheme,
		fleetv1alpha1api.AddToScheme,
		managementv3api.AddToScheme,
		projectv3api.AddToScheme,
		clusterv3api.AddToScheme,
		rkev1api.AddToScheme,
		scheme.AddToScheme,
		apiextensionsv1.AddToScheme,
		apiregistrationv12.AddToScheme,
		catalogv1.AddToScheme,
	}
	AddToScheme = localSchemeBuilder.AddToScheme
	Scheme      = runtime.NewScheme()
)

func init() {
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})
	utilruntime.Must(AddToScheme(Scheme))
	utilruntime.Must(schemes.AddToScheme(Scheme))
}

type Context struct {
	RESTConfig *rest.Config

	Apply               apply.Apply
	Dynamic             *dynamic.Controller
	RKE                 rkecontrollers.Interface
	Mgmt                managementv3.Interface
	Apps                appsv1.Interface
	Admission           admissionregcontrollers.Interface
	Batch               batchv1.Interface
	Fleet               fleetv1alpha1.Interface
	Catalog             catalogcontrollers.Interface
	ControllerFactory   controller.SharedControllerFactory
	MultiClusterManager MultiClusterManager
	TunnelServer        *remotedialer.Server
	Provisioning        provisioningv1.Interface
	RBAC                rbacv1.Interface
	Core                corev1.Interface
	API                 apiregv1.Interface
	CRD                 crdv1.Interface
	K8s                 *kubernetes.Clientset

	ASL                     accesscontrol.AccessSetLookup
	ClientConfig            clientcmd.ClientConfig
	CachedDiscovery         discovery.CachedDiscoveryInterface
	RESTMapper              meta.RESTMapper
	SharedControllerFactory controller.SharedControllerFactory
	leadership              *leader.Manager
	controllerLock          *sync.Mutex

	RESTClientGetter genericclioptions.RESTClientGetter

	mgmt         *management.Factory
	rbac         *rbac.Factory
	ctlg         *catalog.Factory
	adminReg     *admissionreg.Factory
	apps         *apps.Factory
	capi         *capi.Factory
	rke          *rke.Factory
	fleet        *fleet.Factory
	provisioning *provisioning.Factory
	batch        *batch.Factory
	core         *core.Factory

	started bool
	CAPI    v1beta1.Interface
}

type MultiClusterManager interface {
	NormanSchemas() *types.Schemas
	ClusterDialer(clusterID string) func(ctx context.Context, network, address string) (net.Conn, error)
	Start(ctx context.Context) error
	Wait(ctx context.Context)
	Middleware(next http.Handler) http.Handler
	K8sClient(clusterName string) (kubernetes.Interface, error)
}

// OnLeader registers a callback function to be executed when the current context becomes the leader.
// It calls the OnLeader method of the leadership field in the Context struct and passes the provided function as an argument.
func (w *Context) OnLeader(f func(ctx context.Context) error) {
	w.leadership.OnLeader(f)
}

// StartWithTransaction is a method of the Context struct that starts a transaction and executes a provided callback function within the transaction. It returns an error if the callback fails
func (w *Context) StartWithTransaction(ctx context.Context, f func(context.Context) error) (err error) {
	transaction := controller.NewHandlerTransaction(ctx)
	if err := f(transaction); err != nil {
		transaction.Rollback()
		return err
	}

	if err := w.ControllerFactory.SharedCacheFactory().Start(ctx); err != nil {
		transaction.Rollback()
		return err
	}

	w.ControllerFactory.SharedCacheFactory().WaitForCacheSync(ctx)
	transaction.Commit()
	return w.Start(ctx)
}

// Start registers the current context as started and performs necessary initialization steps.
// It acquires a lock using the controllerLock mutex to ensure thread safety.
// If the context is not already started, it calls the Register method of the Dynamic field in the Context struct,
// passing the provided context and SharedControllerFactory as arguments.
// It sets the started field to true.
// It then calls the Start method of the ControllerFactory field, passing the provided context and 50 as arguments.
// It returns an error if there is any.
// It finally calls the Start method of the leadership field in the Context struct, passing the provided context.
func (w *Context) Start(ctx context.Context) error {
	w.controllerLock.Lock()
	defer w.controllerLock.Unlock()

	if !w.started {
		if err := w.Dynamic.Register(ctx, w.SharedControllerFactory); err != nil {
			return err
		}
		w.started = true
	}

	if err := w.ControllerFactory.Start(ctx, 50); err != nil {
		return err
	}
	w.leadership.Start(ctx)
	return nil
}

// WithAgent returns a shallow copy of the Context that has been configured to use a user agent in its
// clients that is the given userAgent appended to "rancher-%s-%s".
func (w *Context) WithAgent(userAgent string) *Context {
	userAgent = fmt.Sprintf("rancher-%s-%s", settings.ServerVersion.Get(), userAgent)
	wContextCopy := *w
	restConfigCopy := &rest.Config{}
	if w.RESTConfig != nil {
		*restConfigCopy = *w.RESTConfig
		restConfigCopy.UserAgent = userAgent
	}
	k8sClientWithAgent, err := kubernetes.NewForConfig(restConfigCopy)
	if err != nil {
		logrus.Debugf("failed to set agent [%s] on k8s client: %v", userAgent, err)
	}
	if err == nil {
		wContextCopy.K8s = k8sClientWithAgent
	}
	applyWithAgent, err := apply.NewForConfig(restConfigCopy)
	if err != nil {
		logrus.Debugf("failed to set agent [%s] on apply client: %v", userAgent, err)
	}
	if err == nil {
		wContextCopy.Apply = applyWithAgent
	}
	wContextCopy.Dynamic = dynamic.New(wContextCopy.K8s.Discovery())
	wContextCopy.RKE = wContextCopy.rke.WithAgent(userAgent).V1()
	wContextCopy.Mgmt = wContextCopy.mgmt.WithAgent(userAgent).V3()
	wContextCopy.Apps = wContextCopy.apps.WithAgent(userAgent).V1()
	wContextCopy.Admission = wContextCopy.adminReg.WithAgent(userAgent).V1()
	wContextCopy.Batch = wContextCopy.batch.WithAgent(userAgent).V1()
	wContextCopy.Fleet = wContextCopy.fleet.WithAgent(userAgent).V1alpha1()
	wContextCopy.Catalog = wContextCopy.ctlg.WithAgent(userAgent).V1()
	wContextCopy.Provisioning = wContextCopy.provisioning.WithAgent(userAgent).V1()
	wContextCopy.RBAC = wContextCopy.rbac.WithAgent(userAgent).V1()
	wContextCopy.Core = wContextCopy.core.WithAgent(userAgent).V1()

	return &wContextCopy
}

func enableProtobuf(cfg *rest.Config) *rest.Config {
	cpy := rest.CopyConfig(cfg)
	cpy.AcceptContentTypes = "application/vnd.kubernetes.protobuf, application/json"
	cpy.ContentType = "application/json"
	return cpy
}

// NewContext creates a new Context with the given parameters. It initializes the required controller factories and other components needed for the context.
func NewContext(ctx context.Context, clientConfig clientcmd.ClientConfig, restConfig *rest.Config) (*Context, error) {
	sharedOpts := GetOptsFromEnv(Management)
	controllerFactory, err := controller.NewSharedControllerFactoryFromConfigWithOptions(enableProtobuf(restConfig), Scheme, sharedOpts)
	if err != nil {
		return nil, err
	}

	opts := &generic.FactoryOptions{
		SharedControllerFactory: controllerFactory,
	}

	apply, err := apply.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	mgmt, err := management.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	apps, err := apps.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	rbac, err := rbac.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	adminReg, err := admissionreg.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	capi, err := capi.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	rke, err := rke.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	fleet, err := fleet.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	provisioning, err := provisioning.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	helm, err := catalog.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	batch, err := batch.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	core, err := core.NewFactoryFromConfigWithOptions(restConfig, opts)
	if err != nil {
		return nil, err
	}

	k8s, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	asl := accesscontrol.NewAccessStore(ctx, true, rbac.Rbac().V1())

	cache := memory.NewMemCacheClient(k8s.Discovery())
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(cache)
	restClientGetter := &SimpleRESTClientGetter{
		ClientConfig:    clientConfig,
		RESTConfig:      restConfig,
		CachedDiscovery: cache,
		RESTMapper:      restMapper,
	}

	wContext := &Context{
		RESTConfig:              restConfig,
		Apply:                   apply,
		SharedControllerFactory: controllerFactory,
		Dynamic:                 dynamic.New(k8s.Discovery()),
		CAPI:                    capi.Cluster().V1beta1(),
		RKE:                     rke.Rke().V1(),
		Mgmt:                    mgmt.Management().V3(),
		Apps:                    apps.Apps().V1(),
		Admission:               adminReg.Admissionregistration().V1(),
		Fleet:                   fleet.Fleet().V1alpha1(),
		Provisioning:            provisioning.Provisioning().V1(),
		Catalog:                 helm.Catalog().V1(),
		Batch:                   batch.Batch().V1(),
		RBAC:                    rbac.Rbac().V1(),
		Core:                    core.Core().V1(),
		K8s:                     k8s,
		ControllerFactory:       controllerFactory,
		ASL:                     asl,
		ClientConfig:            clientConfig,
		MultiClusterManager:     noopMCM{},
		CachedDiscovery:         cache,
		RESTMapper:              restMapper,
		controllerLock:          &sync.Mutex{},
		RESTClientGetter:        restClientGetter,

		mgmt:         mgmt,
		apps:         apps,
		adminReg:     adminReg,
		fleet:        fleet,
		provisioning: provisioning,
		ctlg:         helm,
		batch:        batch,
		core:         core,
		capi:         capi,
		rke:          rke,
		rbac:         rbac,
	}

	return wContext, nil
}

type noopMCM struct {
}

func (n noopMCM) NormanSchemas() *types.Schemas {
	return nil
}

func (n noopMCM) ClusterDialer(clusterID string) func(ctx context.Context, network string, address string) (net.Conn, error) {
	return func(ctx context.Context, network string, address string) (net.Conn, error) {
		return nil, fmt.Errorf("no cluster manager")
	}
}

func (n noopMCM) Wait(ctx context.Context) {
}

func (n noopMCM) Start(ctx context.Context) error {
	return nil
}

func (n noopMCM) Middleware(next http.Handler) http.Handler {
	return next
}

func (n noopMCM) K8sClient(clusterName string) (kubernetes.Interface, error) {
	return nil, nil
}

type SimpleRESTClientGetter struct {
	ClientConfig    clientcmd.ClientConfig
	RESTConfig      *rest.Config
	CachedDiscovery discovery.CachedDiscoveryInterface
	RESTMapper      meta.RESTMapper
}

// ToRawKubeConfigLoader returns the client configuration used for loading the raw kubeconfig.
// It simply returns the ClientConfig field of the SimpleRESTClientGetter.
func (s *SimpleRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return s.ClientConfig
}

// ToRESTConfig returns the RESTConfig stored in the SimpleRESTClientGetter.
// It is a getter method that allows access to the REST configuration.
// The RESTConfig field is a pointer to rest.Config defined in the SimpleRESTClientGetter struct.
// It returns the RESTConfig and a nil error.
func (s *SimpleRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	return s.RESTConfig, nil
}

// ToDiscoveryClient returns the cached discovery client
// It returns the CachedDiscovery field of the SimpleRESTClientGetter struct as a CachedDiscoveryInterface and nil error.
func (s *SimpleRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return s.CachedDiscovery, nil
}

// ToRESTMapper returns the REST mapper associated with the SimpleRESTClientGetter.
// It retrieves the REST mapper from the RESTMapper field of the SimpleRESTClientGetter struct and returns it as the first return value.
// The second return value is always nil since no error is encountered during the retrieval process.
func (s *SimpleRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	return s.RESTMapper, nil
}

// GetOptsFromEnv configures a SharedControllersFactoryOptions using env var and return a pointer to it.
func GetOptsFromEnv(contextType controllerContextType) *controller.SharedControllerFactoryOptions {
	return &controller.SharedControllerFactoryOptions{
		SyncOnlyChangedObjects: syncOnlyChangedObjects(contextType),
	}
}

// syncOnlyChangedObjects returns whether the env var CATTLE_SYNC_ONLY_CHANGED_OBJECTS indicates that controllers for the
// given context type should skip running enqueue if the event triggering the update func is not actual update.
func syncOnlyChangedObjects(option controllerContextType) bool {
	skipUpdate := os.Getenv("CATTLE_SYNC_ONLY_CHANGED_OBJECTS")
	if skipUpdate == "" {
		return false
	}
	parts := strings.Split(skipUpdate, ",")

	for _, part := range parts {
		if controllerContextType(part) == option {
			return true
		}
	}
	return false
}

/*
Copyright 2024 Rancher Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	v1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type RKEControlPlaneHandler func(string, *v1.RKEControlPlane) (*v1.RKEControlPlane, error)

type RKEControlPlaneController interface {
	generic.ControllerMeta
	RKEControlPlaneClient

	OnChange(ctx context.Context, name string, sync RKEControlPlaneHandler)
	OnRemove(ctx context.Context, name string, sync RKEControlPlaneHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() RKEControlPlaneCache
}

type RKEControlPlaneClient interface {
	Create(*v1.RKEControlPlane) (*v1.RKEControlPlane, error)
	Update(*v1.RKEControlPlane) (*v1.RKEControlPlane, error)
	UpdateStatus(*v1.RKEControlPlane) (*v1.RKEControlPlane, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.RKEControlPlane, error)
	List(namespace string, opts metav1.ListOptions) (*v1.RKEControlPlaneList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.RKEControlPlane, err error)
}

type RKEControlPlaneCache interface {
	Get(namespace, name string) (*v1.RKEControlPlane, error)
	List(namespace string, selector labels.Selector) ([]*v1.RKEControlPlane, error)

	AddIndexer(indexName string, indexer RKEControlPlaneIndexer)
	GetByIndex(indexName, key string) ([]*v1.RKEControlPlane, error)
}

type RKEControlPlaneIndexer func(obj *v1.RKEControlPlane) ([]string, error)

type rKEControlPlaneController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewRKEControlPlaneController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) RKEControlPlaneController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &rKEControlPlaneController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromRKEControlPlaneHandlerToHandler(sync RKEControlPlaneHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.RKEControlPlane
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.RKEControlPlane))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *rKEControlPlaneController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.RKEControlPlane))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateRKEControlPlaneDeepCopyOnChange(client RKEControlPlaneClient, obj *v1.RKEControlPlane, handler func(obj *v1.RKEControlPlane) (*v1.RKEControlPlane, error)) (*v1.RKEControlPlane, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *rKEControlPlaneController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *rKEControlPlaneController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *rKEControlPlaneController) OnChange(ctx context.Context, name string, sync RKEControlPlaneHandler) {
	c.AddGenericHandler(ctx, name, FromRKEControlPlaneHandlerToHandler(sync))
}

func (c *rKEControlPlaneController) OnRemove(ctx context.Context, name string, sync RKEControlPlaneHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromRKEControlPlaneHandlerToHandler(sync)))
}

func (c *rKEControlPlaneController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *rKEControlPlaneController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *rKEControlPlaneController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *rKEControlPlaneController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *rKEControlPlaneController) Cache() RKEControlPlaneCache {
	return &rKEControlPlaneCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *rKEControlPlaneController) Create(obj *v1.RKEControlPlane) (*v1.RKEControlPlane, error) {
	result := &v1.RKEControlPlane{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *rKEControlPlaneController) Update(obj *v1.RKEControlPlane) (*v1.RKEControlPlane, error) {
	result := &v1.RKEControlPlane{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *rKEControlPlaneController) UpdateStatus(obj *v1.RKEControlPlane) (*v1.RKEControlPlane, error) {
	result := &v1.RKEControlPlane{}
	return result, c.client.UpdateStatus(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *rKEControlPlaneController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *rKEControlPlaneController) Get(namespace, name string, options metav1.GetOptions) (*v1.RKEControlPlane, error) {
	result := &v1.RKEControlPlane{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *rKEControlPlaneController) List(namespace string, opts metav1.ListOptions) (*v1.RKEControlPlaneList, error) {
	result := &v1.RKEControlPlaneList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *rKEControlPlaneController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *rKEControlPlaneController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.RKEControlPlane, error) {
	result := &v1.RKEControlPlane{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type rKEControlPlaneCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *rKEControlPlaneCache) Get(namespace, name string) (*v1.RKEControlPlane, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1.RKEControlPlane), nil
}

func (c *rKEControlPlaneCache) List(namespace string, selector labels.Selector) (ret []*v1.RKEControlPlane, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.RKEControlPlane))
	})

	return ret, err
}

func (c *rKEControlPlaneCache) AddIndexer(indexName string, indexer RKEControlPlaneIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.RKEControlPlane))
		},
	}))
}

func (c *rKEControlPlaneCache) GetByIndex(indexName, key string) (result []*v1.RKEControlPlane, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.RKEControlPlane, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.RKEControlPlane))
	}
	return result, nil
}

type RKEControlPlaneStatusHandler func(obj *v1.RKEControlPlane, status v1.RKEControlPlaneStatus) (v1.RKEControlPlaneStatus, error)

type RKEControlPlaneGeneratingHandler func(obj *v1.RKEControlPlane, status v1.RKEControlPlaneStatus) ([]runtime.Object, v1.RKEControlPlaneStatus, error)

func RegisterRKEControlPlaneStatusHandler(ctx context.Context, controller RKEControlPlaneController, condition condition.Cond, name string, handler RKEControlPlaneStatusHandler) {
	statusHandler := &rKEControlPlaneStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromRKEControlPlaneHandlerToHandler(statusHandler.sync))
}

func RegisterRKEControlPlaneGeneratingHandler(ctx context.Context, controller RKEControlPlaneController, apply apply.Apply,
	condition condition.Cond, name string, handler RKEControlPlaneGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &rKEControlPlaneGeneratingHandler{
		RKEControlPlaneGeneratingHandler: handler,
		apply:                            apply,
		name:                             name,
		gvk:                              controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterRKEControlPlaneStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type rKEControlPlaneStatusHandler struct {
	client    RKEControlPlaneClient
	condition condition.Cond
	handler   RKEControlPlaneStatusHandler
}

func (a *rKEControlPlaneStatusHandler) sync(key string, obj *v1.RKEControlPlane) (*v1.RKEControlPlane, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type rKEControlPlaneGeneratingHandler struct {
	RKEControlPlaneGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *rKEControlPlaneGeneratingHandler) Remove(key string, obj *v1.RKEControlPlane) (*v1.RKEControlPlane, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.RKEControlPlane{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *rKEControlPlaneGeneratingHandler) Handle(obj *v1.RKEControlPlane, status v1.RKEControlPlaneStatus) (v1.RKEControlPlaneStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.RKEControlPlaneGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}

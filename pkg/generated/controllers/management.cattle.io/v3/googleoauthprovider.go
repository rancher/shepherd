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

package v3

import (
	"context"
	"time"

	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/generic"
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

type GoogleOAuthProviderHandler func(string, *v3.GoogleOAuthProvider) (*v3.GoogleOAuthProvider, error)

type GoogleOAuthProviderController interface {
	generic.ControllerMeta
	GoogleOAuthProviderClient

	OnChange(ctx context.Context, name string, sync GoogleOAuthProviderHandler)
	OnRemove(ctx context.Context, name string, sync GoogleOAuthProviderHandler)
	Enqueue(name string)
	EnqueueAfter(name string, duration time.Duration)

	Cache() GoogleOAuthProviderCache
}

type GoogleOAuthProviderClient interface {
	Create(*v3.GoogleOAuthProvider) (*v3.GoogleOAuthProvider, error)
	Update(*v3.GoogleOAuthProvider) (*v3.GoogleOAuthProvider, error)

	Delete(name string, options *metav1.DeleteOptions) error
	Get(name string, options metav1.GetOptions) (*v3.GoogleOAuthProvider, error)
	List(opts metav1.ListOptions) (*v3.GoogleOAuthProviderList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v3.GoogleOAuthProvider, err error)
}

type GoogleOAuthProviderCache interface {
	Get(name string) (*v3.GoogleOAuthProvider, error)
	List(selector labels.Selector) ([]*v3.GoogleOAuthProvider, error)

	AddIndexer(indexName string, indexer GoogleOAuthProviderIndexer)
	GetByIndex(indexName, key string) ([]*v3.GoogleOAuthProvider, error)
}

type GoogleOAuthProviderIndexer func(obj *v3.GoogleOAuthProvider) ([]string, error)

type googleOAuthProviderController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewGoogleOAuthProviderController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) GoogleOAuthProviderController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &googleOAuthProviderController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromGoogleOAuthProviderHandlerToHandler(sync GoogleOAuthProviderHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v3.GoogleOAuthProvider
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v3.GoogleOAuthProvider))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *googleOAuthProviderController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v3.GoogleOAuthProvider))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateGoogleOAuthProviderDeepCopyOnChange(client GoogleOAuthProviderClient, obj *v3.GoogleOAuthProvider, handler func(obj *v3.GoogleOAuthProvider) (*v3.GoogleOAuthProvider, error)) (*v3.GoogleOAuthProvider, error) {
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

func (c *googleOAuthProviderController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *googleOAuthProviderController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *googleOAuthProviderController) OnChange(ctx context.Context, name string, sync GoogleOAuthProviderHandler) {
	c.AddGenericHandler(ctx, name, FromGoogleOAuthProviderHandlerToHandler(sync))
}

func (c *googleOAuthProviderController) OnRemove(ctx context.Context, name string, sync GoogleOAuthProviderHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromGoogleOAuthProviderHandlerToHandler(sync)))
}

func (c *googleOAuthProviderController) Enqueue(name string) {
	c.controller.Enqueue("", name)
}

func (c *googleOAuthProviderController) EnqueueAfter(name string, duration time.Duration) {
	c.controller.EnqueueAfter("", name, duration)
}

func (c *googleOAuthProviderController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *googleOAuthProviderController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *googleOAuthProviderController) Cache() GoogleOAuthProviderCache {
	return &googleOAuthProviderCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *googleOAuthProviderController) Create(obj *v3.GoogleOAuthProvider) (*v3.GoogleOAuthProvider, error) {
	result := &v3.GoogleOAuthProvider{}
	return result, c.client.Create(context.TODO(), "", obj, result, metav1.CreateOptions{})
}

func (c *googleOAuthProviderController) Update(obj *v3.GoogleOAuthProvider) (*v3.GoogleOAuthProvider, error) {
	result := &v3.GoogleOAuthProvider{}
	return result, c.client.Update(context.TODO(), "", obj, result, metav1.UpdateOptions{})
}

func (c *googleOAuthProviderController) Delete(name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), "", name, *options)
}

func (c *googleOAuthProviderController) Get(name string, options metav1.GetOptions) (*v3.GoogleOAuthProvider, error) {
	result := &v3.GoogleOAuthProvider{}
	return result, c.client.Get(context.TODO(), "", name, result, options)
}

func (c *googleOAuthProviderController) List(opts metav1.ListOptions) (*v3.GoogleOAuthProviderList, error) {
	result := &v3.GoogleOAuthProviderList{}
	return result, c.client.List(context.TODO(), "", result, opts)
}

func (c *googleOAuthProviderController) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), "", opts)
}

func (c *googleOAuthProviderController) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v3.GoogleOAuthProvider, error) {
	result := &v3.GoogleOAuthProvider{}
	return result, c.client.Patch(context.TODO(), "", name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type googleOAuthProviderCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *googleOAuthProviderCache) Get(name string) (*v3.GoogleOAuthProvider, error) {
	obj, exists, err := c.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v3.GoogleOAuthProvider), nil
}

func (c *googleOAuthProviderCache) List(selector labels.Selector) (ret []*v3.GoogleOAuthProvider, err error) {

	err = cache.ListAll(c.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v3.GoogleOAuthProvider))
	})

	return ret, err
}

func (c *googleOAuthProviderCache) AddIndexer(indexName string, indexer GoogleOAuthProviderIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v3.GoogleOAuthProvider))
		},
	}))
}

func (c *googleOAuthProviderCache) GetByIndex(indexName, key string) (result []*v3.GoogleOAuthProvider, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v3.GoogleOAuthProvider, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v3.GoogleOAuthProvider))
	}
	return result, nil
}

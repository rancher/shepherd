/*
Copyright 2025 Rancher Labs, Inc.

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
	"sync"
	"time"

	"github.com/rancher/shepherd/pkg/wrangler/pkg/generic"
	"github.com/rancher/wrangler/v2/pkg/apply"
	"github.com/rancher/wrangler/v2/pkg/condition"
	"github.com/rancher/wrangler/v2/pkg/kv"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StatefulSetController interface for managing StatefulSet resources.
type StatefulSetController interface {
	generic.ControllerInterface[*v1.StatefulSet, *v1.StatefulSetList]
}

// StatefulSetClient interface for managing StatefulSet resources in Kubernetes.
type StatefulSetClient interface {
	generic.ClientInterface[*v1.StatefulSet, *v1.StatefulSetList]
}

// StatefulSetCache interface for retrieving StatefulSet resources in memory.
type StatefulSetCache interface {
	generic.CacheInterface[*v1.StatefulSet]
}

// StatefulSetStatusHandler is executed for every added or modified StatefulSet. Should return the new status to be updated
type StatefulSetStatusHandler func(obj *v1.StatefulSet, status v1.StatefulSetStatus) (v1.StatefulSetStatus, error)

// StatefulSetGeneratingHandler is the top-level handler that is executed for every StatefulSet event. It extends StatefulSetStatusHandler by a returning a slice of child objects to be passed to apply.Apply
type StatefulSetGeneratingHandler func(obj *v1.StatefulSet, status v1.StatefulSetStatus) ([]runtime.Object, v1.StatefulSetStatus, error)

// RegisterStatefulSetStatusHandler configures a StatefulSetController to execute a StatefulSetStatusHandler for every events observed.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterStatefulSetStatusHandler(ctx context.Context, controller StatefulSetController, condition condition.Cond, name string, handler StatefulSetStatusHandler) {
	statusHandler := &statefulSetStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, generic.FromObjectHandlerToHandler(statusHandler.sync))
}

// RegisterStatefulSetGeneratingHandler configures a StatefulSetController to execute a StatefulSetGeneratingHandler for every events observed, passing the returned objects to the provided apply.Apply.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterStatefulSetGeneratingHandler(ctx context.Context, controller StatefulSetController, apply apply.Apply,
	condition condition.Cond, name string, handler StatefulSetGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &statefulSetGeneratingHandler{
		StatefulSetGeneratingHandler: handler,
		apply:                        apply,
		name:                         name,
		gvk:                          controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterStatefulSetStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type statefulSetStatusHandler struct {
	client    StatefulSetClient
	condition condition.Cond
	handler   StatefulSetStatusHandler
}

// sync is executed on every resource addition or modification. Executes the configured handlers and sends the updated status to the Kubernetes API
func (a *statefulSetStatusHandler) sync(key string, obj *v1.StatefulSet) (*v1.StatefulSet, error) {
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

type statefulSetGeneratingHandler struct {
	StatefulSetGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
	seen  sync.Map
}

// Remove handles the observed deletion of a resource, cascade deleting every associated resource previously applied
func (a *statefulSetGeneratingHandler) Remove(key string, obj *v1.StatefulSet) (*v1.StatefulSet, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.StatefulSet{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	if a.opts.UniqueApplyForResourceVersion {
		a.seen.Delete(key)
	}

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

// Handle executes the configured StatefulSetGeneratingHandler and pass the resulting objects to apply.Apply, finally returning the new status of the resource
func (a *statefulSetGeneratingHandler) Handle(obj *v1.StatefulSet, status v1.StatefulSetStatus) (v1.StatefulSetStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.StatefulSetGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}
	if !a.isNewResourceVersion(obj) {
		return newStatus, nil
	}

	err = generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
	if err != nil {
		return newStatus, err
	}
	a.storeResourceVersion(obj)
	return newStatus, nil
}

// isNewResourceVersion detects if a specific resource version was already successfully processed.
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *statefulSetGeneratingHandler) isNewResourceVersion(obj *v1.StatefulSet) bool {
	if !a.opts.UniqueApplyForResourceVersion {
		return true
	}

	// Apply once per resource version
	key := obj.Namespace + "/" + obj.Name
	previous, ok := a.seen.Load(key)
	return !ok || previous != obj.ResourceVersion
}

// storeResourceVersion keeps track of the latest resource version of an object for which Apply was executed
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *statefulSetGeneratingHandler) storeResourceVersion(obj *v1.StatefulSet) {
	if !a.opts.UniqueApplyForResourceVersion {
		return
	}

	key := obj.Namespace + "/" + obj.Name
	a.seen.Store(key, obj.ResourceVersion)
}

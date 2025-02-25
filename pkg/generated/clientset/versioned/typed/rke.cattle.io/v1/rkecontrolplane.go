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

	v1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	scheme "github.com/rancher/shepherd/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// RKEControlPlanesGetter has a method to return a RKEControlPlaneInterface.
// A group's client should implement this interface.
type RKEControlPlanesGetter interface {
	RKEControlPlanes(namespace string) RKEControlPlaneInterface
}

// RKEControlPlaneInterface has methods to work with RKEControlPlane resources.
type RKEControlPlaneInterface interface {
	Create(ctx context.Context, rKEControlPlane *v1.RKEControlPlane, opts metav1.CreateOptions) (*v1.RKEControlPlane, error)
	Update(ctx context.Context, rKEControlPlane *v1.RKEControlPlane, opts metav1.UpdateOptions) (*v1.RKEControlPlane, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, rKEControlPlane *v1.RKEControlPlane, opts metav1.UpdateOptions) (*v1.RKEControlPlane, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.RKEControlPlane, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.RKEControlPlaneList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.RKEControlPlane, err error)
	RKEControlPlaneExpansion
}

// rKEControlPlanes implements RKEControlPlaneInterface
type rKEControlPlanes struct {
	*gentype.ClientWithList[*v1.RKEControlPlane, *v1.RKEControlPlaneList]
}

// newRKEControlPlanes returns a RKEControlPlanes
func newRKEControlPlanes(c *RkeV1Client, namespace string) *rKEControlPlanes {
	return &rKEControlPlanes{
		gentype.NewClientWithList[*v1.RKEControlPlane, *v1.RKEControlPlaneList](
			"rkecontrolplanes",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.RKEControlPlane { return &v1.RKEControlPlane{} },
			func() *v1.RKEControlPlaneList { return &v1.RKEControlPlaneList{} }),
	}
}

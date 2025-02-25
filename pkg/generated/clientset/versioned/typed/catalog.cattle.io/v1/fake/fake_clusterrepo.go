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

package fake

import (
	"context"

	v1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeClusterRepos implements ClusterRepoInterface
type FakeClusterRepos struct {
	Fake *FakeCatalogV1
}

var clusterreposResource = v1.SchemeGroupVersion.WithResource("clusterrepos")

var clusterreposKind = v1.SchemeGroupVersion.WithKind("ClusterRepo")

// Get takes name of the clusterRepo, and returns the corresponding clusterRepo object, and an error if there is any.
func (c *FakeClusterRepos) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ClusterRepo, err error) {
	emptyResult := &v1.ClusterRepo{}
	obj, err := c.Fake.
		Invokes(testing.NewRootGetActionWithOptions(clusterreposResource, name, options), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ClusterRepo), err
}

// List takes label and field selectors, and returns the list of ClusterRepos that match those selectors.
func (c *FakeClusterRepos) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ClusterRepoList, err error) {
	emptyResult := &v1.ClusterRepoList{}
	obj, err := c.Fake.
		Invokes(testing.NewRootListActionWithOptions(clusterreposResource, clusterreposKind, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ClusterRepoList{ListMeta: obj.(*v1.ClusterRepoList).ListMeta}
	for _, item := range obj.(*v1.ClusterRepoList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested clusterRepos.
func (c *FakeClusterRepos) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchActionWithOptions(clusterreposResource, opts))
}

// Create takes the representation of a clusterRepo and creates it.  Returns the server's representation of the clusterRepo, and an error, if there is any.
func (c *FakeClusterRepos) Create(ctx context.Context, clusterRepo *v1.ClusterRepo, opts metav1.CreateOptions) (result *v1.ClusterRepo, err error) {
	emptyResult := &v1.ClusterRepo{}
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateActionWithOptions(clusterreposResource, clusterRepo, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ClusterRepo), err
}

// Update takes the representation of a clusterRepo and updates it. Returns the server's representation of the clusterRepo, and an error, if there is any.
func (c *FakeClusterRepos) Update(ctx context.Context, clusterRepo *v1.ClusterRepo, opts metav1.UpdateOptions) (result *v1.ClusterRepo, err error) {
	emptyResult := &v1.ClusterRepo{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateActionWithOptions(clusterreposResource, clusterRepo, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ClusterRepo), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeClusterRepos) UpdateStatus(ctx context.Context, clusterRepo *v1.ClusterRepo, opts metav1.UpdateOptions) (result *v1.ClusterRepo, err error) {
	emptyResult := &v1.ClusterRepo{}
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceActionWithOptions(clusterreposResource, "status", clusterRepo, opts), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ClusterRepo), err
}

// Delete takes name of the clusterRepo and deletes it. Returns an error if one occurs.
func (c *FakeClusterRepos) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteActionWithOptions(clusterreposResource, name, opts), &v1.ClusterRepo{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeClusterRepos) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewRootDeleteCollectionActionWithOptions(clusterreposResource, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ClusterRepoList{})
	return err
}

// Patch applies the patch and returns the patched clusterRepo.
func (c *FakeClusterRepos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ClusterRepo, err error) {
	emptyResult := &v1.ClusterRepo{}
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceActionWithOptions(clusterreposResource, name, pt, data, opts, subresources...), emptyResult)
	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ClusterRepo), err
}

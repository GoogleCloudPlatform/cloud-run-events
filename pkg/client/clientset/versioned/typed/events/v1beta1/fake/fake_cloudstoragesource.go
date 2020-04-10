/*
Copyright 2020 Google LLC

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1beta1 "github.com/google/knative-gcp/pkg/apis/events/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCloudStorageSources implements CloudStorageSourceInterface
type FakeCloudStorageSources struct {
	Fake *FakeEventsV1beta1
	ns   string
}

var cloudstoragesourcesResource = schema.GroupVersionResource{Group: "events.cloud.google.com", Version: "v1beta1", Resource: "cloudstoragesources"}

var cloudstoragesourcesKind = schema.GroupVersionKind{Group: "events.cloud.google.com", Version: "v1beta1", Kind: "CloudStorageSource"}

// Get takes name of the cloudStorageSource, and returns the corresponding cloudStorageSource object, and an error if there is any.
func (c *FakeCloudStorageSources) Get(name string, options v1.GetOptions) (result *v1beta1.CloudStorageSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cloudstoragesourcesResource, c.ns, name), &v1beta1.CloudStorageSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CloudStorageSource), err
}

// List takes label and field selectors, and returns the list of CloudStorageSources that match those selectors.
func (c *FakeCloudStorageSources) List(opts v1.ListOptions) (result *v1beta1.CloudStorageSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cloudstoragesourcesResource, cloudstoragesourcesKind, c.ns, opts), &v1beta1.CloudStorageSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.CloudStorageSourceList{ListMeta: obj.(*v1beta1.CloudStorageSourceList).ListMeta}
	for _, item := range obj.(*v1beta1.CloudStorageSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cloudStorageSources.
func (c *FakeCloudStorageSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cloudstoragesourcesResource, c.ns, opts))

}

// Create takes the representation of a cloudStorageSource and creates it.  Returns the server's representation of the cloudStorageSource, and an error, if there is any.
func (c *FakeCloudStorageSources) Create(cloudStorageSource *v1beta1.CloudStorageSource) (result *v1beta1.CloudStorageSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cloudstoragesourcesResource, c.ns, cloudStorageSource), &v1beta1.CloudStorageSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CloudStorageSource), err
}

// Update takes the representation of a cloudStorageSource and updates it. Returns the server's representation of the cloudStorageSource, and an error, if there is any.
func (c *FakeCloudStorageSources) Update(cloudStorageSource *v1beta1.CloudStorageSource) (result *v1beta1.CloudStorageSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cloudstoragesourcesResource, c.ns, cloudStorageSource), &v1beta1.CloudStorageSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CloudStorageSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCloudStorageSources) UpdateStatus(cloudStorageSource *v1beta1.CloudStorageSource) (*v1beta1.CloudStorageSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(cloudstoragesourcesResource, "status", c.ns, cloudStorageSource), &v1beta1.CloudStorageSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CloudStorageSource), err
}

// Delete takes name of the cloudStorageSource and deletes it. Returns an error if one occurs.
func (c *FakeCloudStorageSources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cloudstoragesourcesResource, c.ns, name), &v1beta1.CloudStorageSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCloudStorageSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cloudstoragesourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.CloudStorageSourceList{})
	return err
}

// Patch applies the patch and returns the patched cloudStorageSource.
func (c *FakeCloudStorageSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.CloudStorageSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cloudstoragesourcesResource, c.ns, name, pt, data, subresources...), &v1beta1.CloudStorageSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.CloudStorageSource), err
}

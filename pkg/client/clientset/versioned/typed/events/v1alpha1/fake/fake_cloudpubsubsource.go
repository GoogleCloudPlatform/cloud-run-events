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
	v1alpha1 "github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeCloudPubSubSources implements CloudPubSubSourceInterface
type FakeCloudPubSubSources struct {
	Fake *FakeEventsV1alpha1
	ns   string
}

var cloudpubsubsourcesResource = schema.GroupVersionResource{Group: "events.cloud.google.com", Version: "v1alpha1", Resource: "cloudpubsubsources"}

var cloudpubsubsourcesKind = schema.GroupVersionKind{Group: "events.cloud.google.com", Version: "v1alpha1", Kind: "CloudPubSubSource"}

// Get takes name of the cloudPubSubSource, and returns the corresponding cloudPubSubSource object, and an error if there is any.
func (c *FakeCloudPubSubSources) Get(name string, options v1.GetOptions) (result *v1alpha1.CloudPubSubSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(cloudpubsubsourcesResource, c.ns, name), &v1alpha1.CloudPubSubSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CloudPubSubSource), err
}

// List takes label and field selectors, and returns the list of CloudPubSubSources that match those selectors.
func (c *FakeCloudPubSubSources) List(opts v1.ListOptions) (result *v1alpha1.CloudPubSubSourceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(cloudpubsubsourcesResource, cloudpubsubsourcesKind, c.ns, opts), &v1alpha1.CloudPubSubSourceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.CloudPubSubSourceList{ListMeta: obj.(*v1alpha1.CloudPubSubSourceList).ListMeta}
	for _, item := range obj.(*v1alpha1.CloudPubSubSourceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested cloudPubSubSources.
func (c *FakeCloudPubSubSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(cloudpubsubsourcesResource, c.ns, opts))

}

// Create takes the representation of a cloudPubSubSource and creates it.  Returns the server's representation of the cloudPubSubSource, and an error, if there is any.
func (c *FakeCloudPubSubSources) Create(cloudPubSubSource *v1alpha1.CloudPubSubSource) (result *v1alpha1.CloudPubSubSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(cloudpubsubsourcesResource, c.ns, cloudPubSubSource), &v1alpha1.CloudPubSubSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CloudPubSubSource), err
}

// Update takes the representation of a cloudPubSubSource and updates it. Returns the server's representation of the cloudPubSubSource, and an error, if there is any.
func (c *FakeCloudPubSubSources) Update(cloudPubSubSource *v1alpha1.CloudPubSubSource) (result *v1alpha1.CloudPubSubSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(cloudpubsubsourcesResource, c.ns, cloudPubSubSource), &v1alpha1.CloudPubSubSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CloudPubSubSource), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeCloudPubSubSources) UpdateStatus(cloudPubSubSource *v1alpha1.CloudPubSubSource) (*v1alpha1.CloudPubSubSource, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(cloudpubsubsourcesResource, "status", c.ns, cloudPubSubSource), &v1alpha1.CloudPubSubSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CloudPubSubSource), err
}

// Delete takes name of the cloudPubSubSource and deletes it. Returns an error if one occurs.
func (c *FakeCloudPubSubSources) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(cloudpubsubsourcesResource, c.ns, name), &v1alpha1.CloudPubSubSource{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCloudPubSubSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(cloudpubsubsourcesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.CloudPubSubSourceList{})
	return err
}

// Patch applies the patch and returns the patched cloudPubSubSource.
func (c *FakeCloudPubSubSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CloudPubSubSource, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(cloudpubsubsourcesResource, c.ns, name, pt, data, subresources...), &v1alpha1.CloudPubSubSource{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.CloudPubSubSource), err
}

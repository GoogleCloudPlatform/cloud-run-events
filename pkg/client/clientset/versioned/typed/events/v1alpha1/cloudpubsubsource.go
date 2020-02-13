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

package v1alpha1

import (
	"time"

	v1alpha1 "github.com/google/knative-gcp/pkg/apis/events/v1alpha1"
	scheme "github.com/google/knative-gcp/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// CloudPubSubSourcesGetter has a method to return a CloudPubSubSourceInterface.
// A group's client should implement this interface.
type CloudPubSubSourcesGetter interface {
	CloudPubSubSources(namespace string) CloudPubSubSourceInterface
}

// CloudPubSubSourceInterface has methods to work with CloudPubSubSource resources.
type CloudPubSubSourceInterface interface {
	Create(*v1alpha1.CloudPubSubSource) (*v1alpha1.CloudPubSubSource, error)
	Update(*v1alpha1.CloudPubSubSource) (*v1alpha1.CloudPubSubSource, error)
	UpdateStatus(*v1alpha1.CloudPubSubSource) (*v1alpha1.CloudPubSubSource, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.CloudPubSubSource, error)
	List(opts v1.ListOptions) (*v1alpha1.CloudPubSubSourceList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CloudPubSubSource, err error)
	CloudPubSubSourceExpansion
}

// cloudPubSubSources implements CloudPubSubSourceInterface
type cloudPubSubSources struct {
	client rest.Interface
	ns     string
}

// newCloudPubSubSources returns a CloudPubSubSources
func newCloudPubSubSources(c *EventsV1alpha1Client, namespace string) *cloudPubSubSources {
	return &cloudPubSubSources{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the cloudPubSubSource, and returns the corresponding cloudPubSubSource object, and an error if there is any.
func (c *cloudPubSubSources) Get(name string, options v1.GetOptions) (result *v1alpha1.CloudPubSubSource, err error) {
	result = &v1alpha1.CloudPubSubSource{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CloudPubSubSources that match those selectors.
func (c *cloudPubSubSources) List(opts v1.ListOptions) (result *v1alpha1.CloudPubSubSourceList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.CloudPubSubSourceList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested cloudPubSubSources.
func (c *cloudPubSubSources) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a cloudPubSubSource and creates it.  Returns the server's representation of the cloudPubSubSource, and an error, if there is any.
func (c *cloudPubSubSources) Create(cloudPubSubSource *v1alpha1.CloudPubSubSource) (result *v1alpha1.CloudPubSubSource, err error) {
	result = &v1alpha1.CloudPubSubSource{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		Body(cloudPubSubSource).
		Do().
		Into(result)
	return
}

// Update takes the representation of a cloudPubSubSource and updates it. Returns the server's representation of the cloudPubSubSource, and an error, if there is any.
func (c *cloudPubSubSources) Update(cloudPubSubSource *v1alpha1.CloudPubSubSource) (result *v1alpha1.CloudPubSubSource, err error) {
	result = &v1alpha1.CloudPubSubSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		Name(cloudPubSubSource.Name).
		Body(cloudPubSubSource).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *cloudPubSubSources) UpdateStatus(cloudPubSubSource *v1alpha1.CloudPubSubSource) (result *v1alpha1.CloudPubSubSource, err error) {
	result = &v1alpha1.CloudPubSubSource{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		Name(cloudPubSubSource.Name).
		SubResource("status").
		Body(cloudPubSubSource).
		Do().
		Into(result)
	return
}

// Delete takes name of the cloudPubSubSource and deletes it. Returns an error if one occurs.
func (c *cloudPubSubSources) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *cloudPubSubSources) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched cloudPubSubSource.
func (c *cloudPubSubSources) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.CloudPubSubSource, err error) {
	result = &v1alpha1.CloudPubSubSource{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("cloudpubsubsources").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

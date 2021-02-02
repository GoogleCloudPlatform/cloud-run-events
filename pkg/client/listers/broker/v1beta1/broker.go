/*
Copyright 2021 Google LLC

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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/google/knative-gcp/pkg/apis/broker/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// BrokerLister helps list Brokers.
// All objects returned here must be treated as read-only.
type BrokerLister interface {
	// List lists all Brokers in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.Broker, err error)
	// Brokers returns an object that can list and get Brokers.
	Brokers(namespace string) BrokerNamespaceLister
	BrokerListerExpansion
}

// brokerLister implements the BrokerLister interface.
type brokerLister struct {
	indexer cache.Indexer
}

// NewBrokerLister returns a new BrokerLister.
func NewBrokerLister(indexer cache.Indexer) BrokerLister {
	return &brokerLister{indexer: indexer}
}

// List lists all Brokers in the indexer.
func (s *brokerLister) List(selector labels.Selector) (ret []*v1beta1.Broker, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.Broker))
	})
	return ret, err
}

// Brokers returns an object that can list and get Brokers.
func (s *brokerLister) Brokers(namespace string) BrokerNamespaceLister {
	return brokerNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// BrokerNamespaceLister helps list and get Brokers.
// All objects returned here must be treated as read-only.
type BrokerNamespaceLister interface {
	// List lists all Brokers in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.Broker, err error)
	// Get retrieves the Broker from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.Broker, error)
	BrokerNamespaceListerExpansion
}

// brokerNamespaceLister implements the BrokerNamespaceLister
// interface.
type brokerNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Brokers in the indexer for a given namespace.
func (s brokerNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.Broker, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.Broker))
	})
	return ret, err
}

// Get retrieves the Broker from the indexer for a given namespace and name.
func (s brokerNamespaceLister) Get(name string) (*v1beta1.Broker, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("broker"), name)
	}
	return obj.(*v1beta1.Broker), nil
}

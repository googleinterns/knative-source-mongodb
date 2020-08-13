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

package testing

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"

	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
)

// MongoDbSourceOption enables further configuration of a CronJob.
type MongoDbSourceOption func(*v1alpha1.MongoDbSource)

// NewMongoDbSource creates a MongoDbSource with MongoDbSourceOption.
func NewMongoDbSource(name, namespace string, o ...MongoDbSourceOption) *v1alpha1.MongoDbSource {
	c := &v1alpha1.MongoDbSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range o {
		opt(c)
	}
	c.SetDefaults(context.Background()) // TODO: We should add defaults and validation.
	return c
}

// WithMongoDbSourceUID initializes the UID of the MongoDbSource.
func WithMongoDbSourceUID(uid string) MongoDbSourceOption {
	return func(c *v1alpha1.MongoDbSource) {
		c.UID = types.UID(uid)
	}
}

// WithInitMongoDbSourceConditions initializes the MongoDbSource's conditions.
func WithInitMongoDbSourceConditions(s *v1alpha1.MongoDbSource) {
	s.Status.InitializeConditions()
}

// WithMongoDbSourceSpec returns a function that updates the Spec field of the MongoDbSource.
func WithMongoDbSourceSpec(spec v1alpha1.MongoDbSourceSpec) MongoDbSourceOption {
	return func(c *v1alpha1.MongoDbSource) {
		c.Spec = spec
	}
}

// WithMongoDbSourceSinkNotFound updates the status of the sink to be not found.
func WithMongoDbSourceSinkNotFound(s *v1alpha1.MongoDbSource) {
	s.Status.MarkNoSink("NotFound", "")
}

// WithMongoDbSourceSink updates the status of the sink to be found.
func WithMongoDbSourceSink(uri *apis.URL) MongoDbSourceOption {
	return func(s *v1alpha1.MongoDbSource) {
		s.Status.MarkSink(uri)
	}
}

// WithMongoDbSourceNotDeployed updates the status of the source to Not Deployed.
func WithMongoDbSourceNotDeployed(name string) MongoDbSourceOption {
	return func(s *v1alpha1.MongoDbSource) {
		s.Status.PropagateDeploymentAvailability(NewDeployment(name, "any"))
	}
}

// WithMongoDbSourceDeployed updates the status of the source to Deployed.
func WithMongoDbSourceDeployed(s *v1alpha1.MongoDbSource) {
	s.Status.PropagateDeploymentAvailability(NewDeployment("any", "any", WithDeploymentAvailable()))
}

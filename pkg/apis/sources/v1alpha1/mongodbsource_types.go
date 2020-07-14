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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	"knative.dev/pkg/webhook/resourcesemantics"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MongoDbSource is the Schema for the MongoDbSource API.
// +k8s:openapi-gen=true
type MongoDbSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDbSourceSpec   `json:"spec"`
	Status MongoDbSourceStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (m *MongoDbSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("MongoDbSource")
}

var (
	// Check that MongoDbSource can be validated and defaulted.
	_ apis.Validatable = (*MongoDbSource)(nil)
	_ apis.Defaultable = (*MongoDbSource)(nil)
	// Check that we can create OwnerReferences to a MongoDbSource.
	_ kmeta.OwnerRefable = (*MongoDbSource)(nil)
	// Check that MongoDbSource is a runtime.Object.
	_ runtime.Object = (*MongoDbSource)(nil)
	// Check that MongoDbSource satisfies resourcesemantics.GenericCRD.
	_ resourcesemantics.GenericCRD = (*MongoDbSource)(nil)
	// Check that MongoDbSource implements the Conditions duck type.
	_ = duck.VerifyType(&MongoDbSource{}, &duckv1.Conditions{})
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*MongoDbSource)(nil)
)

// MongoDbSourceSpec defines the desired state of MongoDbSource.
type MongoDbSourceSpec struct {
	// ServiceAccountName holds the name of the Kubernetes service account
	// as which the underlying K8s resources should be run. If unspecified
	// this will default to the "default" service account for the namespace
	// in which the MongoDbSource exists.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// MongoDbCredentials is the credential to use to access MongoDb.
	// Must be a secret. Only Name and Namespace are used.
	Secret corev1.LocalObjectReference `json:"secret"`

	// Database is the database to watch for changes.
	Database string `json:"database"`

	// Collection is the collection to watch for changes.
	// +optional
	Collection string `json:"collection,omitempty"`

	// SourceSpec
	// inherits duck/v1 SourceSpec, which currently provides:
	// * Sink - a reference to an object that will resolve to a domain name or
	//   a URI directly to use as the sink.
	// * CloudEventOverrides - defines overrides to control the output format
	//   and modifications of the event sent to the sink.
	duckv1.SourceSpec `json:",inline"`
}

// MongoDbSourceStatus defines the observed state of MongoDbSource.
type MongoDbSourceStatus struct {
	// inherits duck/v1 SourceStatus, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last
	//   processed by the controller.
	// * Conditions - the latest available observations of a resource's current
	//   state.
	// * SinkURI - the current active sink URI that has been configured for the
	//   Source.
	duckv1.SourceStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MongoDbSourceList contains a list of MongoDbSource.
type MongoDbSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDbSource `json:"items"`
}

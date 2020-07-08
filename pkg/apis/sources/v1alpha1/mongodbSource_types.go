/*
Copyright 2019 The Knative Authors

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
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MongodbSource is the Schema for the githubsources API
// +k8s:openapi-gen=true
type MongodbSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongodbSourceSpec   `json:"spec,omitempty"`
	Status MongodbSourceStatus `json:"status,omitempty"`
}

// Check that CouchDb source can be validated and can be defaulted.
var _ runtime.Object = (*MongodbSource)(nil)

// Check that we can create OwnerReferences to a Configuration.
var _ kmeta.OwnerRefable = (*MongodbSource)(nil)

// Check that the type conforms to the duck Knative Resource shape.
var _ duckv1.KRShaped = (*MongodbSource)(nil)

// Check that MongodbSource implements the Conditions duck type.
var _ = duck.VerifyType(&MongodbSource{}, &duckv1.Conditions{})

// FeedType is the type of Feed
type FeedType string

var MongodbSourceEventTypes = []string{
	MongodbSourceUpdateEventType,
	MongodbSourceDeleteEventType,
}

const (
	// MongodbSourceUpdateEventType is the MongodbSource CloudEvent type for update.
	MongodbSourceUpdateEventType = "org.apache.couchdb.document.update"

	// MongodbSourceDeleteEventType is the MongodbSource CloudEvent type for deletion.
	MongodbSourceDeleteEventType = "org.apache.couchdb.document.delete"

	// FeedNormal corresponds to the "normal" feed. The connection to the server
	// is closed after reporting changes.
	FeedNormal = FeedType("normal")

	// FeedContinuous corresponds to the "continuous" feed. The connection to the
	// server stays open after reporting changes.
	FeedContinuous = FeedType("continuous")
)

// MongodbSourceSpec defines the desired state of MongodbSource
type MongodbSourceSpec struct {
	// ServiceAccountName holds the name of the Kubernetes service account
	// as which the underlying K8s resources should be run. If unspecified
	// this will default to the "default" service account for the namespace
	// in which the MongodbSource exists.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// CouchDbCredentials is the credential to use to access CouchDb.
	// Must be a secret. Only Name and Namespace are used.
	CouchDbCredentials corev1.ObjectReference `json:"credentials,omitempty"`

	// Feed changes how CouchDB sends the response.
	// More information: https://docs.couchdb.org/en/stable/api/database/changes.html#changes-feeds
	Feed FeedType `json:"feed"`

	// Database is the database to watch for changes
	Database string `json:"database"`

	// Sink is a reference to an object that will resolve to a domain name to use as the sink.
	// +optional
	Sink *duckv1.Destination `json:"sink,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (s *MongodbSource) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("MongodbSource")
}

// MongodbSourceStatus defines the observed state of MongodbSource
type MongodbSourceStatus struct {
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

// MongodbSourceList contains a list of MongodbSource
type MongodbSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongodbSource `json:"items"`
}

// GetStatus retrieves the duck status for this resource. Implements the KRShaped interface.
func (c *MongodbSource) GetStatus() *duckv1.Status {
	return &c.Status.Status
}

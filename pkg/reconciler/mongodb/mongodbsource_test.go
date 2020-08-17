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

package mongodb

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/resolver"

	mongowrapper "github.com/googleinterns/knative-source-mongodb/pkg/mongo"
	"k8s.io/client-go/kubernetes/scheme"
	clientgotesting "k8s.io/client-go/testing"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	logtesting "knative.dev/pkg/logging/testing"

	fakesourcesclient "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/client/fake"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"

	sourcesv1alpha1 "github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	"github.com/googleinterns/knative-source-mongodb/pkg/client/injection/reconciler/sources/v1alpha1/mongodbsource"
	. "github.com/googleinterns/knative-source-mongodb/pkg/reconciler/testing"
	"knative.dev/pkg/client/injection/ducks/duck/v1/addressable"
	. "knative.dev/pkg/reconciler/testing"
)

var (
	sinkDNS = sinkName + ".testnamespace.svc.cluster.local"
	sinkURI = &apis.URL{
		Scheme: "http",
		Host:   sinkDNS,
		Path:   "/",
	}
)

const (
	testRAImage = "github.com/googleinterns/knative-source-mongodb/test/image"
	sourceName  = "test-mongodb-source"
	sourceUID   = "1234"
	testNS      = "testnamespace"
	sinkName    = "testsink"
	secretName  = "test-secret"
	db          = "db"
	coll        = "coll"
)

func init() {
	// Add types to scheme
	_ = sourcesv1alpha1.AddToScheme(scheme.Scheme)
}

func TestAllCases(t *testing.T) {
	table := TableTest{{
		Name: "bad workqueue key",
		// Make sure Reconcile handles bad keys.
		Key: "too/many/parts",
	}, {
		Name: "key not found",
		// Make sure Reconcile handles good keys that don't exist.
		Key: "foo/not-found",
	}, {
		Name:    "missing sink",
		WantErr: true,
		Objects: []runtime.Object{
			NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					Secret: corev1.LocalObjectReference{
						Name: "secret",
					},
				}),
				WithMongoDbSourceUID(sourceUID),
			),
		},
		Key: testNS + "/" + sourceName,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					Secret: corev1.LocalObjectReference{
						Name: "secret",
					},
				}),
				WithMongoDbSourceUID(sourceUID),
				// Status Update:
				WithInitMongoDbSourceConditions,
				WithMongoDbSourceSinkNotFound,
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, `UpdateFailed Failed to update status for "test-mongodb-source":`,
				`missing field(s): spec.sink`),
		},
	}, {
		Name:    "missing secret",
		WantErr: true,
		Objects: []runtime.Object{
			NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					SourceSpec: duckv1.SourceSpec{Sink: newSinkDestination()},
				}),
				WithMongoDbSourceUID(sourceUID),
			),
			newSink(),
		},
		Key: testNS + "/" + sourceName,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					SourceSpec: duckv1.SourceSpec{Sink: newSinkDestination()},
				}),
				WithMongoDbSourceUID(sourceUID),
				// Status Update:
				WithInitMongoDbSourceConditions,
				WithMongoDbSourceSink(sinkURI),
				WithMongoDbSourceConnectionFailed(`secret "" not found`),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, `UpdateFailed Failed to update status for "test-mongodb-source":`,
				`missing field(s): spec.secret`),
		},
	}, {
		Name:    "secret has no URI field",
		WantErr: true,
		Objects: []runtime.Object{
			NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					Secret: corev1.LocalObjectReference{
						Name: secretName,
					},
					SourceSpec: duckv1.SourceSpec{Sink: newSinkDestination()},
				}),
				WithMongoDbSourceUID(sourceUID),
			),
			newSink(),
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNS,
				},
				Data: map[string][]byte{
					"notURI": []byte("secretURI"),
				},
			},
		},
		Key: testNS + "/" + sourceName,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					Secret: corev1.LocalObjectReference{
						Name: secretName,
					},
					SourceSpec: duckv1.SourceSpec{Sink: newSinkDestination()},
				}),
				WithMongoDbSourceUID(sourceUID),
				// Status Update:
				WithInitMongoDbSourceConditions,
				WithMongoDbSourceSink(sinkURI),
				WithMongoDbSourceConnectionFailed("Unable to get MongoDb URI field"),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError",
				"Unable to get MongoDb URI field"),
		},
	}, {
		Name:    "can't create mongodb client: does't start with mongodb://",
		WantErr: true,
		Objects: []runtime.Object{
			NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					Secret: corev1.LocalObjectReference{
						Name: secretName,
					},
					SourceSpec: duckv1.SourceSpec{Sink: newSinkDestination()},
				}),
				WithMongoDbSourceUID(sourceUID),
			),
			newSink(),
			&corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretName,
					Namespace: testNS,
				},
				Data: map[string][]byte{
					"URI": []byte("notValid"),
				},
			},
		},
		Key: testNS + "/" + sourceName,
		WantStatusUpdates: []clientgotesting.UpdateActionImpl{{
			Object: NewMongoDbSource(sourceName, testNS,
				WithMongoDbSourceSpec(sourcesv1alpha1.MongoDbSourceSpec{
					Database:   db,
					Collection: coll,
					Secret: corev1.LocalObjectReference{
						Name: secretName,
					},
					SourceSpec: duckv1.SourceSpec{Sink: newSinkDestination()},
				}),
				WithMongoDbSourceUID(sourceUID),
				// Status Update:
				WithInitMongoDbSourceConditions,
				WithMongoDbSourceSink(sinkURI),
				WithMongoDbSourceConnectionFailed(`error parsing uri: scheme must be "mongodb" or "mongodb+srv"`),
			),
		}},
		WantEvents: []string{
			Eventf(corev1.EventTypeWarning, "InternalError",
				`error parsing uri: scheme must be "mongodb" or "mongodb+srv"`),
		},
	},
	}

	defer logtesting.ClearAll()
	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher, testData map[string]interface{}) controller.Reconciler {
		ctx = addressable.WithDuck(ctx)
		r := &Reconciler{
			kubeClientSet:       fakekubeclient.Get(ctx),
			deploymentLister:    listers.GetDeploymentLister(),
			secretLister:        listers.GetSecretLister(),
			receiveAdapterImage: testRAImage,
			configs:             &reconcilersource.EmptyVarsGenerator{},
			sinkResolver:        resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
			createClientFn:      mongowrapper.NewClient,
		}

		return mongodbsource.NewReconciler(ctx, logging.FromContext(ctx), fakesourcesclient.Get(ctx), listers.GetMongoDbSourceLister(), controller.GetEventRecorder(ctx), r)
	}))
}

// newSink returns an unstructured v1.Service which is special-cased for resolving the URI.
func newSink() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"namespace": testNS,
				"name":      sinkName,
			},
		},
	}
}

// newSinkDestination returns an unstructured v1.Service which is special-cased for resolving the URI.
func newSinkDestination() duckv1.Destination {
	return duckv1.Destination{
		Ref: &duckv1.KReference{
			APIVersion: "v1",
			Kind:       "Service",
			Name:       sinkName,
			Namespace:  testNS,
		},
	}
}

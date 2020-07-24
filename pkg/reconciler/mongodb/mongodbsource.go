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

package mongodbsource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"knative.dev/pkg/logging"

	v1alpha1 "github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	mongodbsource "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/reconciler/sources/v1alpha1/mongodbsource"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	corev1 "k8s.io/api/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	reconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason MongoDbSourceReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "MongoDbSourceReconciled", "MongoDbSource reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements controller.Reconciler for MongoDbSource resources.
type Reconciler struct {
	// Lister
	secretLister corev1listers.SecretLister

	sinkResolver *resolver.URIResolver
}

// Check that our Reconciler implements Interface
var _ mongodbsource.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind. Reconciles based on secret, credentials,
// database & collection presence, sink resolvability and creates corresponding receive adapter.
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.MongoDbSource) reconciler.Event {
	src.Status.InitializeConditions()
	src.Status.ObservedGeneration = src.Generation

	// Check the secret, credentials, database and collection existance.
	err := r.checkConnection(ctx, src)
	if err != nil {
		src.Status.MarkConnectionFailed(err)
		return err
	}
	src.Status.MarkConnectionSuccess()

	// Check the resolvability of the specified sink
	sinkURI, err := r.resolveSink(ctx, src)
	if err != nil {
		src.Status.MarkNoSink("NotFound", "")
		return err
	}
	src.Status.MarkSink(sinkURI)

	// r.reconcileReceiveAdapter(ctx, src)
	// TODO: make rsc.make receive adapter
	// https://github.com/vaikas/postgressource/blob/b116b1097b87b9855a711f085f22996c522027bb/pkg/reconciler/deployment.go#L42
	//  with lister, if doesnt exist create deployment
	// if exist check hwo its different

	return newReconciledNormal(src.Namespace, src.Name)
}

// checkConnection checks the secret, credentials, database and collection existance.
func (r *Reconciler) checkConnection(ctx context.Context, src *v1alpha1.MongoDbSource) reconciler.Event {
	// Try to connect to the database and see if it works.
	secret, err := r.secretLister.Secrets(src.Namespace).Get(src.Spec.Secret.Name)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to read MongoDb credentials secret", zap.Error(err))
		return err
	}
	rawURI, ok := secret.Data["URI"]
	if !ok {
		logging.FromContext(ctx).Desugar().Error("Unable to get MongoDb URI field", zap.Any("secretName", secret.Name), zap.Any("secretNamespace", secret.Namespace))
		return err
	}
	URI := string(rawURI)

	// Connect to the MongoDb replica-set.
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Couldn't connect to database", zap.Error(err))
		return err
	}
	err = client.Connect(ctx)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Couldn't connect to database", zap.Error(err))
		return err
	}
	defer client.Disconnect(ctx)

	// See if database exists in available databases.
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		logging.FromContext(ctx).Error("Couldn't look up existing databases", zap.Error(err))
		return err
	}
	if !stringInSlice(src.Spec.Database, databases) {
		err = errors.New("Couldn't find database name in available databases")
		logging.FromContext(ctx).Error("Couldn't find database in available database", zap.Any("database", src.Spec.Database), zap.Any("availableDatabases", fmt.Sprint(databases)), zap.Error(err))
		return err
	}

	// See if collection exists in available collections.
	if src.Spec.Collection != "" {
		collections, err := client.Database(src.Spec.Database).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			logging.FromContext(ctx).Error("Couldn't look up existing collections", zap.Error(err))
			return err
		}
		if !stringInSlice(src.Spec.Collection, collections) {
			err = errors.New("Couldn't find collection name in available collections")
			logging.FromContext(ctx).Error("Couldn't find collection in available collections", zap.Any("collection", src.Spec.Collection), zap.Any("availableCollections", fmt.Sprint(collections)), zap.Error(err))
			return err
		}
	}

	return nil
}

// checkSink checks the resolvability of the specified sink
func (r *Reconciler) resolveSink(ctx context.Context, src *v1alpha1.MongoDbSource) (*apis.URL, reconciler.Event) {
	dest := src.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		if dest.Ref.Namespace == "" {
			dest.Ref.Namespace = src.Namespace
		}
	}

	sinkURI, err := r.sinkResolver.URIFromDestinationV1(*dest, src)
	if err != nil {
		return nil, newWarningSinkNotFound(dest)
	}

	return sinkURI, nil
}

// Helper Fucntion: Creates Event about Sink not being found.
func newWarningSinkNotFound(sink *duckv1.Destination) reconciler.Event {
	b, _ := json.Marshal(sink)
	return reconciler.NewEvent(corev1.EventTypeWarning, "SinkNotFound", "Sink not found: %s", string(b))
}

// Helper function: finds if string exists in array of strings.
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

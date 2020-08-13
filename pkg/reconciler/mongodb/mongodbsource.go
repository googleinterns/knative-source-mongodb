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
	"fmt"
	"net/url"

	"go.uber.org/zap"
	"knative.dev/pkg/logging"

	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	"github.com/googleinterns/knative-source-mongodb/pkg/client/injection/reconciler/sources/v1alpha1/mongodbsource"
	"github.com/googleinterns/knative-source-mongodb/pkg/reconciler/mongodb/resources"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"
)

// Reconciler implements controller.Reconciler for MongoDbSource resources.
type Reconciler struct {
	receiveAdapterImage string `envconfig:"MONGODB_RA_IMAGE" required:"true"`
	kubeClientSet       kubernetes.Interface
	sinkResolver        *resolver.URIResolver

	// Lister
	deploymentLister appsv1listers.DeploymentLister
	secretLister     corev1listers.SecretLister

	configs reconcilersource.ConfigAccessor
}

// Check that our Reconciler implements Interface
var _ mongodbsource.Interface = (*Reconciler)(nil)

// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, src *v1alpha1.MongoDbSource) reconciler.Event {
	// Steps:
	// 1. Resolve the sink.
	// 2. Ensure it can connect to the DB with the specified credentials, and that the DB and collection exists.
	// 3. Reconcile the receive adapter.

	// Resolve the specified sink.
	sinkURI, err := r.resolveSink(ctx, src)
	if err != nil {
		src.Status.MarkNoSink("NotFound", "")
		return err
	}
	src.Status.MarkSink(sinkURI)

	// Check that we can connect to the DB.
	err = r.checkConnection(ctx, src)
	if err != nil {
		src.Status.MarkConnectionFailed(err)
		return err
	}
	src.Status.MarkConnectionSuccess()

	// Reconcile the receive adapter.
	ra, err := r.reconcileReceiveAdapter(ctx, src)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Failed to reconcile Deployment", zap.Error(err))
		return err
	}
	src.Status.PropagateDeploymentAvailability(ra)

	return nil
}

// checkConnection checks the secret, credentials, database and collection existence.
func (r *Reconciler) checkConnection(ctx context.Context, src *v1alpha1.MongoDbSource) reconciler.Event {
	// Try to connect to the database and see if it works.
	secret, err := r.secretLister.Secrets(src.Namespace).Get(src.Spec.Secret.Name)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Unable to read MongoDb credentials secret", zap.Error(err))
		return err
	}
	rawURI, ok := secret.Data["URI"]
	if !ok {
		logging.FromContext(ctx).Desugar().Error("Unable to get MongoDb URI field", zap.Any("secret", secret.Name))
		return err
	}
	URI := string(rawURI)

	// Connect to the MongoDb replica-set.
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Error creating mongo client", zap.Error(err))
		return err
	}
	err = client.Connect(ctx)
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Error connecting to database", zap.Error(err))
		return err
	}
	defer client.Disconnect(ctx)

	// See if database exists in available databases.
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		logging.FromContext(ctx).Desugar().Error("Error listing databases", zap.Error(err))
		return err
	}
	if !stringInSlice(src.Spec.Database, databases) {
		err = fmt.Errorf("database %q not available in existing databases", src.Spec.Database)
		logging.FromContext(ctx).Desugar().Error("Database not available in existing databases", zap.Any("database", src.Spec.Database), zap.Any("availableDatabases", fmt.Sprint(databases)), zap.Error(err))
		return err
	}

	// See if collection exists in available collections.
	if src.Spec.Collection != "" {
		collections, err := client.Database(src.Spec.Database).ListCollectionNames(ctx, bson.M{})
		if err != nil {
			logging.FromContext(ctx).Desugar().Error("Error listing collections", zap.Error(err))
			return err
		}
		if !stringInSlice(src.Spec.Collection, collections) {
			err = fmt.Errorf("collection %q not available in existing collections", src.Spec.Collection)
			logging.FromContext(ctx).Desugar().Error("Collection not available in existing collections", zap.Any("collection", src.Spec.Collection), zap.Any("availableCollections", fmt.Sprint(collections)), zap.Error(err))
			return err
		}
	}

	return nil
}

// resolveSink checks the resolvability of the specified sink.
func (r *Reconciler) resolveSink(ctx context.Context, src *v1alpha1.MongoDbSource) (*apis.URL, error) {
	dest := src.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		if dest.Ref.Namespace == "" {
			dest.Ref.Namespace = src.Namespace
		}
	}

	return r.sinkResolver.URIFromDestinationV1(*dest, src)
}

// reconcileReceiveAdapter reconciles the Receive Adapter Deployment.
func (r *Reconciler) reconcileReceiveAdapter(ctx context.Context, src *v1alpha1.MongoDbSource) (*appsv1.Deployment, error) {
	ceSourcePrefix, err := r.makeCeSourcePrefix(ctx, src)
	args := &resources.ReceiveAdapterArgs{
		Image:          r.receiveAdapterImage,
		Labels:         resources.Labels(src.Name),
		Source:         src,
		CeSourcePrefix: ceSourcePrefix,
		SinkURL:        src.Status.SinkURI.String(),
		Configs:        r.configs,
	}
	expected, err := resources.MakeReceiveAdapter(args)
	if err != nil {
		return nil, err
	}
	ra, err := r.deploymentLister.Deployments(expected.Namespace).Get(expected.Name)
	if apierrors.IsNotFound(err) {
		ra, err = r.kubeClientSet.AppsV1().Deployments(expected.Namespace).Create(expected)
		if err != nil {
			return nil, err
		}
		return ra, nil
	} else if err != nil {
		return nil, fmt.Errorf("error getting receive adapter %q: %v", expected.Name, err)
	} else if !metav1.IsControlledBy(ra, src.GetObjectMeta()) {
		return nil, fmt.Errorf("deployment %q is not owned by %s %q",
			ra.Name, src.GetGroupVersionKind().Kind, src.GetObjectMeta().GetName())
	} else if r.podSpecImageSync(expected.Spec.Template.Spec, ra.Spec.Template.Spec) {
		if ra, err = r.kubeClientSet.AppsV1().Deployments(expected.Namespace).Update(ra); err != nil {
			return ra, err
		}
		return ra, nil
	} else {
		logging.FromContext(ctx).Desugar().Debug("Reusing existing receive adapter", zap.Any("receiveAdapter", ra))
	}
	return ra, nil
}

// Returns false if an update is needed.
func (r *Reconciler) podSpecImageSync(expected corev1.PodSpec, now corev1.PodSpec) bool {
	// got needs all of the containers that want as, but it is allowed to have more.
	dirty := false
	for _, ec := range expected.Containers {
		n, nc := getContainer(ec.Name, now)
		if nc == nil {
			now.Containers = append(now.Containers, ec)
			dirty = true
			continue
		}
		if nc.Image != ec.Image {
			now.Containers[n].Image = ec.Image
			dirty = true
		}
	}
	return dirty
}

// getContainer gets a container by name.
func getContainer(name string, spec corev1.PodSpec) (int, *corev1.Container) {
	for i, c := range spec.Containers {
		if c.Name == name {
			return i, &c
		}
	}
	return -1, nil
}

// makeCeSourcePrefix computes the Cloud Event source prefix for the Event Source variable.
func (r *Reconciler) makeCeSourcePrefix(ctx context.Context, src *v1alpha1.MongoDbSource) (string, error) {
	secret, err := r.secretLister.Secrets(src.Namespace).Get(src.Spec.Secret.Name)
	if err != nil {
		logging.FromContext(ctx).Error("Unable to read MongoDb credentials secret", zap.Error(err))
		return "", err
	}
	rawURI, ok := secret.Data["URI"]
	if !ok {
		logging.FromContext(ctx).Error("Unable to get MongoDb URI field", zap.Any("secretName", secret.Name), zap.Any("secretNamespace", secret.Namespace))
		return "", err
	}

	url, err := url.Parse(string(rawURI))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("mongodb://%s", url.Hostname()), nil
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

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
	"os"

	mongowrapper "github.com/googleinterns/knative-source-mongodb/pkg/mongo"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"

	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	"github.com/googleinterns/knative-source-mongodb/pkg/client/injection/informers/sources/v1alpha1/mongodbsource"
	v1alpha1mongodbsource "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/reconciler/sources/v1alpha1/mongodbsource"
	"k8s.io/client-go/tools/cache"
	reconcilersource "knative.dev/eventing/pkg/reconciler/source"
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"
)

// Declare Constants.
const (
	// raImageEnvVar is the name of the environment variable that contains the receive adapter's
	// image. It must be defined.
	raImageEnvVar = "MONGODB_RA_IMAGE"

	component = "mongodbsource"
)

// NewController creates a Reconciler for MongoDbSource and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	// Setup Event informers.
	deploymentInformer := deploymentinformer.Get(ctx)
	mongodbsourceInformer := mongodbsource.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	raImage, defined := os.LookupEnv(raImageEnvVar)
	if !defined {
		logging.FromContext(ctx).Errorf("required environment variable %q not defined", raImageEnvVar)
		return nil
	}

	r := &Reconciler{
		receiveAdapterImage: raImage,
		kubeClientSet:       kubeclient.Get(ctx),
		secretLister:        secretInformer.Lister(),
		deploymentLister:    deploymentInformer.Lister(),
		configs:             reconcilersource.WatchConfigurations(ctx, component, cmw),
		createClientFn:      mongowrapper.NewClient,
	}
	impl := v1alpha1mongodbsource.NewImpl(ctx, r)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	mongoGK := v1alpha1.Kind("MongoDbSource")

	// Set up the event handlers.
	logger.Info("Setting up event handlers.")
	mongodbsourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(mongoGK),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(mongoGK),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})
	return impl
}

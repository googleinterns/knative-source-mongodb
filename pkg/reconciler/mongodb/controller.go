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
	context "context"

	v1alpha1 "github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	mongodbsource "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/informers/sources/v1alpha1/mongodbsource"
	v1alpha1mongodbsource "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/reconciler/sources/v1alpha1/mongodbsource"
	"k8s.io/client-go/tools/cache"
	secretinformer "knative.dev/pkg/client/injection/kube/informers/core/v1/secret"
	configmap "knative.dev/pkg/configmap"
	controller "knative.dev/pkg/controller"
	logging "knative.dev/pkg/logging"
	"knative.dev/pkg/resolver"

)

// NewController creates a Reconciler for MongoDbSource and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	// Setup Event informers.
	mongodbsourceInformer := mongodbsource.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	r := &Reconciler{
		secretLister: secretInformer.Lister(),
	}
	impl := v1alpha1mongodbsource.NewImpl(ctx, r)

	r.sinkResolver = resolver.NewURIResolver(ctx, impl.EnqueueKey)

	// Set up the event handlers.
	logger.Info("Setting up event handlers.")
	mongodbsourceInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))
	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterGroupKind(v1alpha1.Kind("MongoDbSource")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}

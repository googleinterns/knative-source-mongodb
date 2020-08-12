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
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/resolver"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
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

const (
	testRAImage = "github.com/googleinterns/knative-source-mongodb/test/image"
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
	}}

	// TODO add more UTs

	defer logtesting.ClearAll()
	table.Test(t, MakeFactory(func(ctx context.Context, listers *Listers, cmw configmap.Watcher, testData map[string]interface{}) controller.Reconciler {
		ctx = addressable.WithDuck(ctx)
		r := &Reconciler{
			kubeClientSet:       fakekubeclient.Get(ctx),
			deploymentLister:    listers.GetDeploymentLister(),
			receiveAdapterImage: testRAImage,
			configs:             &reconcilersource.EmptyVarsGenerator{},
			sinkResolver:        resolver.NewURIResolver(ctx, func(types.NamespacedName) {}),
		}

		return mongodbsource.NewReconciler(ctx, logging.FromContext(ctx), fakesourcesclient.Get(ctx), listers.GetMongoDbSourceLister(), controller.GetEventRecorder(ctx), r)
	}))

}

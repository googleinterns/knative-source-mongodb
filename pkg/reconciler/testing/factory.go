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
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"knative.dev/pkg/reconciler"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	logtesting "knative.dev/pkg/logging/testing"

	fakesourcesclient "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/client/fake"
	fakekubeclient "knative.dev/pkg/client/injection/kube/client/fake"
	fakedynamicclient "knative.dev/pkg/injection/clients/dynamicclient/fake"

	. "knative.dev/pkg/reconciler/testing"
)

const (
	// maxEventBufferSize is the estimated max number of event notifications that
	// can be buffered during reconciliation.
	maxEventBufferSize = 10
)

// Ctor functions create a k8s controller with given params.
type Ctor func(context.Context, *Listers, configmap.Watcher, map[string]interface{}) controller.Reconciler

// MakeFactory creates a reconciler factory with fake clients and controller created by `ctor`.
func MakeFactory(ctor Ctor) Factory {
	return func(t *testing.T, r *TableRow) (controller.Reconciler, ActionRecorderList, EventList) {
		ls := NewListers(r.Objects)

		ctx := context.Background()
		logger := logtesting.TestLogger(t)
		ctx = logging.WithLogger(ctx, logger)

		ctx, kubeClient := fakekubeclient.With(ctx, ls.GetKubeObjects()...)
		ctx, client := fakesourcesclient.With(ctx, ls.GetSourcesObjects()...)

		dynamicScheme := runtime.NewScheme()
		for _, addTo := range clientSetSchemes {
			_ = addTo(dynamicScheme)
		}
		ctx, dynamicClient := fakedynamicclient.With(ctx, dynamicScheme, ls.GetAllObjects()...)

		// The dynamic client's support for patching is BS.  Implement it
		// here via PrependReactor (this can be overridden below by the
		// provided reactors).
		dynamicClient.PrependReactor("patch", "*",
			func(action ktesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			})

		eventRecorder := record.NewFakeRecorder(maxEventBufferSize)
		ctx = controller.WithEventRecorder(ctx, eventRecorder)

		PrependGenerateNameReactor(&client.Fake)
		PrependGenerateNameReactor(&dynamicClient.Fake)

		// Set up our Controller from the fakes.
		c := ctor(ctx, &ls, configmap.NewStaticWatcher(), r.OtherTestData)

		// If the reconcilers is leader aware, then promote it.
		if la, ok := c.(reconciler.LeaderAware); ok {
			la.Promote(reconciler.UniversalBucket(), func(reconciler.Bucket, types.NamespacedName) {})
		}

		for _, reactor := range r.WithReactors {
			kubeClient.PrependReactor("*", "*", reactor)
			client.PrependReactor("*", "*", reactor)
			dynamicClient.PrependReactor("*", "*", reactor)
		}

		// Validate all Create operations through the serving client.
		client.PrependReactor("create", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return ValidateCreates(ctx, action)
		})
		client.PrependReactor("update", "*", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
			return ValidateUpdates(ctx, action)
		})

		actionRecorderList := ActionRecorderList{dynamicClient, client, kubeClient}
		eventList := EventList{Recorder: eventRecorder}

		return c, actionRecorderList, eventList

	}
}

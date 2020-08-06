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
	"testing"

	"knative.dev/pkg/configmap"
	logtesting "knative.dev/pkg/logging/testing"
	. "knative.dev/pkg/reconciler/testing"

	// Fake injection informers
	_ "github.com/googleinterns/knative-source-mongodb/pkg/client/clientset/versioned/typed/sources/v1alpha1/fake"
	_ "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/client/fake"
	_ "github.com/googleinterns/knative-source-mongodb/pkg/client/injection/informers/sources/v1alpha1/mongodbsource/fake"
	_ "github.com/googleinterns/knative-source-mongodb/pkg/reconciler/testing"
	_ "knative.dev/pkg/client/injection/kube/informers/batch/v1/job/fake"
	_ "knative.dev/pkg/client/injection/kube/informers/core/v1/serviceaccount/fake"
)

func TestNew(t *testing.T) {
	defer logtesting.ClearAll()
	ctx, _ := SetupFakeContext(t)
	cmw := configmap.NewStaticWatcher()
	c := NewController(ctx, cmw)
	if c == nil {
		t.Fatal("Expected NewController to return a non-nil value")
	}
}

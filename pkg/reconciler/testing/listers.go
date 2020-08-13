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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	fakesourcesclientset "github.com/googleinterns/knative-source-mongodb/pkg/client/clientset/versioned/fake"
	v1alpha1listers "github.com/googleinterns/knative-source-mongodb/pkg/client/listers/sources/v1alpha1"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	"knative.dev/pkg/reconciler/testing"
)

var sinkAddToScheme = func(scheme *runtime.Scheme) error {
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: "testing.google.com", Version: "v1alpha1", Kind: "Sink"}, &unstructured.Unstructured{})
	return nil
}

var clientSetSchemes = []func(*runtime.Scheme) error{
	fakekubeclientset.AddToScheme,
	fakesourcesclientset.AddToScheme,
	sinkAddToScheme,
}

// Listers holds sorters.
type Listers struct {
	sorter testing.ObjectSorter
}

// NewListers creates a Listers object with objs.
func NewListers(objs []runtime.Object) Listers {
	scheme := runtime.NewScheme()

	for _, addTo := range clientSetSchemes {
		addTo(scheme)
	}

	ls := Listers{
		sorter: testing.NewObjectSorter(scheme),
	}

	ls.sorter.AddObjects(objs...)

	return ls
}

// indexerFor returns the indexer for the given object.
func (l *Listers) indexerFor(obj runtime.Object) cache.Indexer {
	return l.sorter.IndexerForObjectType(obj)
}

// GetKubeObjects returns the kube objects.
func (l *Listers) GetKubeObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakekubeclientset.AddToScheme)
}

// GetSourcesObjects returns the sources objects.
func (l *Listers) GetSourcesObjects() []runtime.Object {
	return l.sorter.ObjectsForSchemeFunc(fakesourcesclientset.AddToScheme)
}

// GetAllObjects returns all the objects.
func (l *Listers) GetAllObjects() []runtime.Object {
	all := l.GetSourcesObjects()
	all = append(all, l.GetKubeObjects()...)
	return all
}

// GetDeploymentLister returns the Deployment lister.
func (l *Listers) GetDeploymentLister() appsv1listers.DeploymentLister {
	return appsv1listers.NewDeploymentLister(l.indexerFor(&appsv1.Deployment{}))
}

// GetSecretLister returns the Secret lister.
func (l *Listers) GetSecretLister() corev1listers.SecretLister {
	return corev1listers.NewSecretLister(l.indexerFor(&corev1.Secret{}))
}

// GetMongoDbSourceLister returns the MongoDbSource lister.
func (l *Listers) GetMongoDbSourceLister() v1alpha1listers.MongoDbSourceLister {
	return v1alpha1listers.NewMongoDbSourceLister(l.indexerFor(&v1alpha1.MongoDbSource{}))
}

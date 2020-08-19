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

	mongo "github.com/googleinterns/knative-source-mongodb/pkg/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// database wraps mongo.Database. It is the client that will be used everywhere except unit tests.
type testDatabase struct {
	data TestDbData
}

// TestDbData is the data used to configure the test MongoDb Database.
type TestDbData struct {
	ListCollErr error
	Collections []string
}

// Verify that it satisfies the mongo.Database interface.
var _ mongo.Database = &testDatabase{}

// ListCollectionNames implements mongo.Client.Database.ListCollectionNames.
func (tdb *testDatabase) ListCollectionNames(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]string, error) {
	if tdb.data.ListCollErr != nil {
		return nil, tdb.data.ListCollErr
	}
	return tdb.data.Collections, nil
}

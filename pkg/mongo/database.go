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

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// database wraps mongo.Database. It is the client that will be used everywhere except unit tests.
type database struct {
	database *mongo.Database
}

// Verify that it satisfies the mongo.Database interface.
var _ Database = &database{}

// ListCollectionNames implements mongo.Client.Database.ListCollectionNames.
func (db *database) ListCollectionNames(ctx context.Context, filter interface{}, opts ...*options.ListCollectionsOptions) ([]string, error) {
	return db.database.ListCollectionNames(ctx, filter, opts...)
}

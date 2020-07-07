/*
Copyright 2019 Google LLC

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

package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Function to watch for changes
func iterateChangeStream(routineCtx context.Context, waitGroup sync.WaitGroup, stream *mongo.ChangeStream) {
	// Close watch stream and close the thread in waitgroup when done
	defer stream.Close(routineCtx)
	defer waitGroup.Done()
	// For each new change recorded
	for stream.Next(routineCtx) {
		var data bson.M
		if err := stream.Decode(&data); err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", data)

	}
}

func main() {
	// Connect to the mongo replica-set
	URI := "mongodb://10.12.0.12:27017" // Add IP:Port of main replica set pod
	client, err := mongo.NewClient(options.Client().ApplyURI(URI))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	// Find the Databases and prints them
	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)

	// Retrieve a specific collection
	collection := client.Database("main-db").Collection("collection1")

	// Watch for changes
	// Create a wait group to be able to watch asynchronously
	var waitGroup sync.WaitGroup

	// Create a watch stream
	collectionStream, err := collection.Watch(ctx, mongo.Pipeline{})
	if err != nil {
		panic(err)
	}

	// Add a thread to the wait group
	waitGroup.Add(1)
	// Create Routine context
	routineCtx, _ := context.WithCancel(context.Background())
	// Run the asynch routine
	go iterateChangeStream(routineCtx, waitGroup, collectionStream)
	// Keep running it in the backgorund and wait until its done
	waitGroup.Wait()
}

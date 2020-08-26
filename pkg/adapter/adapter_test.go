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

package adapter

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	// testcloudclient "github.com/eventing/pkg/kncloudevents/testing/"
	"github.com/google/go-cmp/cmp"
	"github.com/googleinterns/knative-source-mongodb/pkg/apis/sources/v1alpha1"
	mongotesting "github.com/googleinterns/knative-source-mongodb/pkg/mongo/testing"
	"go.mongodb.org/mongo-driver/bson"
	testcloudclient "knative.dev/eventing/pkg/adapter/v2/test"
	"knative.dev/pkg/logging"
)

var (
	db          string = "db"
	coll        string = "coll"
	docID       string = "docID"
	ID          string = "ID"
	CESource    string = fmt.Sprintf("CEPrefix/databases/%s/collections/%s", db, coll)
	CEEventType string = v1alpha1.MongoDbSourceEventTypes["insert"]
)

func TestMakeCloudEvent(t *testing.T) {
	tests := []struct {
		name     string
		a        *mongoDbAdapter
		data     bson.M
		wantErr  bool
		wantCEFn func() *cloudevents.Event
	}{
		{
			name: "error decoding bson",
			a: &mongoDbAdapter{
				namespace:      "namespace",
				ceSourcePrefix: "CEPrefix",
				database:       db,
				collection:     coll,
			},
			data: bson.M{
				"NOTns": &bson.M{
					"missing": "mising",
				},
				"operationType": "insert",
			},
			wantErr: true,
		},
		{
			name: "unrecognizable type of change",
			a: &mongoDbAdapter{
				namespace:      "namespace",
				ceSourcePrefix: "CEPrefix",
				database:       db,
				collection:     coll,
			},
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       ID,
					"clusterTime": "",
				},
				"documentKey": bson.M{
					"_id": docID,
				},
				"fullDocument": bson.M{
					"_id":  docID,
					"key1": "value1",
				},
				"operationType": "NOTvalid",
			},
			wantErr: true,
		},
		{
			name: "Valid",
			a: &mongoDbAdapter{
				namespace:      "namespace",
				ceSourcePrefix: "CEPrefix",
				database:       db,
				collection:     coll,
			},
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       ID,
					"clusterTime": "",
				},
				"documentKey": bson.M{
					"_id": docID,
				},
				"fullDocument": bson.M{
					"_id":  docID,
					"key1": "value1",
				},
				"operationType": "insert",
			},
			wantCEFn: func() *cloudevents.Event {
				return makeCloudEventTest(&bson.M{
					"_id":  docID,
					"key1": "value1",
				})
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receivedCE, err := test.a.makeCloudEvent(test.data)
			if err != nil {
				if !test.wantErr {
					t.Errorf("makeCloudEvent got error %q want error=%v", err, test.wantErr)
				}
			} else {
				if diff := cmp.Diff(test.wantCEFn(), receivedCE); diff != "" {
					t.Errorf("makeCloudEvent got unexpeceted cloudevents.Event (-want +got) %s", diff)
				}
			}
		})
	}
}

func makeCloudEventTest(data *bson.M) *cloudevents.Event {
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(fmt.Sprintf("%x", md5.Sum([]byte(ID))))
	event.SetSource(CESource)
	event.SetType(CEEventType)
	event.SetData(cloudevents.ApplicationJSON, data)
	return &event
}

func TestProcessChanges(t *testing.T) {
	tests := []struct {
		name       string
		testCSdata mongotesting.TestCSData
		wantCE     bool
	}{
		{
			name: "decoder error",

			testCSdata: mongotesting.TestCSData{
				DecodeErr:      errors.New("Error decoding the change stream"),
				NextFnExecuted: true,
			},
			wantCE: false,
		},
		{
			name: "CE Creation error",
			testCSdata: mongotesting.TestCSData{
				NewChange: bson.M{
					"ns": bson.M{
						"coll": coll,
						"db":   db,
					},
					"_id": bson.M{
						"_data":       ID,
						"clusterTime": "",
					},
					"documentKey": bson.M{
						"_id": docID,
					},
					"fullDocument": bson.M{
						"_id":  docID,
						"key1": "value1",
					},
					"operationType": "insert",
				},
				NextFnExecuted: true,
			},
			wantCE: false,
		},
		{
			name: "valid",
			testCSdata: mongotesting.TestCSData{
				NewChange: bson.M{
					"ns": bson.M{
						"coll": coll,
						"db":   db,
					},
					"_id": bson.M{
						"_data":       ID,
						"clusterTime": "",
					},
					"documentKey": bson.M{
						"_id": docID,
					},
					"fullDocument": bson.M{
						"_id":  docID,
						"key1": "value1",
					},
					"operationType": "insert",
				},
				NextFnExecuted: true,
			},
			wantCE: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			ce := testcloudclient.NewTestClient()
			a := mongoDbAdapter{
				namespace:      "namespace",
				ceSourcePrefix: "CEPrefix",
				database:       db,
				collection:     coll,
				ceClient:       ce,
				logger:         logging.FromContext(ctx),
			}
			stream := &mongotesting.TestChangeStream{
				Data: test.testCSdata,
			}
			a.processChanges(ctx, stream)
			if test.wantCE {
				validateSent(t, ce, `{"_id":"docID","key1":"value1"}`)
			}
		})
	}
}

func validateSent(t *testing.T, ce *testcloudclient.TestCloudEventsClient, wantData string) {
	if got := len(ce.Sent()); got != 1 {
		t.Errorf("Expected 1 event to be sent, got %d", got)
	}

	if got := ce.Sent()[0].Data(); string(got) != wantData {
		t.Errorf("Expected %q event to be sent, got %q", wantData, string(got))
	}
}

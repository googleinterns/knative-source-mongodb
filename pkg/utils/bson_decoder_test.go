/*
Copyright 2020 Google LLC.
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
package utils

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	id    = "id"
	db    = "db"
	coll  = "coll"
	docID = "docID"
)

func TestBsonDecoder(t *testing.T) {

	tests := []struct {
		name          string
		data          bson.M
		wantErr       bool
		wantChangeObj *ChangeObject
	}{
		{
			name: "no ns field",
			data: bson.M{
				"NOTns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       "IDofChange",
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
			wantErr: true,
		},
		{
			name: "ns field has no coll field",
			data: bson.M{
				"ns": bson.M{
					"NOTcoll": coll,
					"db":      db,
				},
				"_id": bson.M{
					"_data":       "IDofChange",
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
			wantErr: true,
		},
		{
			name: "no _id field",
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"NOT_id": bson.M{
					"_data":       "IDofChange",
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
			wantErr: true,
		},
		{
			name: "no _data in _id field",
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"NOT_data":    "IDofChange",
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
			wantErr: true,
		},
		{
			name: "no operationType field",
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       "IDofChange",
					"clusterTime": "",
				},
				"documentKey": bson.M{
					"_id": docID,
				},
				"fullDocument": bson.M{
					"_id":  docID,
					"key1": "value1",
				},
				"NOToperationType": "insert",
			},
			wantErr: true,
		},
		{
			name: "no documentKey field in deletion",
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       "IDofChange",
					"clusterTime": "",
				},
				"NOTdocumentKey": bson.M{
					"_id": docID,
				},
				"operationType": "delete",
			},
			wantErr: true,
		},
		{
			name: "no fullDocuement field in insert or update",
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       "IDofChange",
					"clusterTime": "",
				},
				"documentKey": bson.M{
					"_id": docID,
				},
				"NOTfullDocument": bson.M{
					"_id":  docID,
					"key1": "value1",
				},
				"operationType": "update",
			},
			wantErr: true,
		},
		{
			name: "valid",
			data: bson.M{
				"ns": bson.M{
					"coll": coll,
					"db":   db,
				},
				"_id": bson.M{
					"_data":       id,
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
			wantErr: false,
			wantChangeObj: &ChangeObject{
				ID:            id,
				OperationType: "insert",
				Collection:    coll,
				Payload: &bson.M{
					"_id":  docID,
					"key1": "value1",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			receivedObj, err := DecodeChangeBson(test.data)
			if err != nil {
				if !test.wantErr {
					t.Errorf("DecodeBson got error %q want error=%v", err, test.wantErr)
				}
			} else {
				if !reflect.DeepEqual(receivedObj, test.wantChangeObj) {
					t.Errorf("Decoded ChangeObject doesn't match desired ChangeObject. Got=%v Want=%v", receivedObj, test.wantChangeObj)
				}
			}
		})
	}
}

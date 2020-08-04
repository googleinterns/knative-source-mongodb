**This is not an officially supported Google product.**

[![GoDoc](https://godoc.org/github.com/googleinterns/knative-source-mongodb?status.svg)](https://godoc.org/github.com/googleinterns/knative-source-mongodb)
[![Go Report Card](https://goreportcard.com/badge/googleinterns/knative-source-mongodb)](https://goreportcard.com/report/googleinterns/knative-source-mongodb)
[![LICENSE](https://img.shields.io/github/license/googleinterns/knative-source-mongodb.svg)](https://github.com/googleinterns/knative-source-mongodb/blob/master/LICENSE)
[![Build Status](https://travis-ci.org/googleinterns/knative-source-mongodb.svg?branch=master)](https://travis-ci.org/googleinterns/knative-source-mongodb)

# Knative Source - MongoDB

## Development

### Dependencies

1. [go](https://golang.org/doc/install)

### Setup environment

Put the following in a `./bashrc` or `./bashprofile`

```sh
export GOPATH="$HOME/go"
export PATH="${PATH}:${GOPATH}/bin"
```

### Clone to your machine

1. [Fork this repo](https://help.github.com/articles/fork-a-repo/) to your account
2. Clone to your machine

```sh
mkdir -p ${GOPATH}/src/github.com/googleinterns
cd ${GOPATH}/src/github.com/googleinterns
git clone git@github.com:${YOUR_GITHUB_USERNAME}/knative-source-mongodb.git
cd knative-source-mongodb
git remote add upstream https://github.com/googleinterns/knative-source-mongodb.git
git remote set-url --push upstream no_push
```

`Upstream` will allow for you to [sync your fork](https://help.github.com/articles/syncing-a-fork/)

### Building

`go build ./...`

### Running unit tests

`go test ./...`

### Updating dependencies

`./hack/update-deps.sh`

## Source Code Headers

Every file containing source code must include copyright and license
information. This includes any JS/CSS files that you might be serving out to
browsers. (This is to help well-intentioned people avoid accidental copying that
doesn't comply with the license.)

Apache header:

    Copyright 2020 Google LLC

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.


# MongoDb source for Knative

The MongoDb Event source adds support of MongoDb ressources to Knative Eventing.

## Usage steps

1. Setup [Knative Eventing](../DEVELOPMENT.md) in your Kubernetes Cluster.
2. Install MongoDb and create a Replica Set.
3. Create a secret containing the data needed to access your MongoDb service.
   For example:

   ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
    name: my-mongo-secret
    namespace: default
    stringData:
            URI: mongodb://USERNAME:PASSWORD@IP:PORT/USERDB
   ```
   where the IP is the IP of the main/principal pod of your replica set. USERDB is the database your user account partains to (can be `admin`).

4. Create the `MongoDbSource` custom objects, by configuring the required
   `database` has to be provided, but `collection` is optional
   For example:

   ```yaml
    apiVersion: sources.google.com/v1alpha1
    kind: MongoDbSource
    metadata:
    name: mongodb-example-source
    namespace: default
    spec:
    database: db1
    collection: coll1  # optional
    secret:
        name: my-mongo-secret
    sink:
        ref:
        apiVersion: serving.knative.dev/v1
        kind: Service
        name: event-display
        namespace: mongodb
   ```

# Installing MongoDb 
## Install the MongoDB Enterprise Kubernetes Operator
1. Create a namespace for your Kubernetes deployment.

`kubectl create namespace mongodb`

2. Create Custom Resource Definitions for MongoDB, MongoDBUser and MongoDBOpsManager (cluster admin permissions required):

`kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-enterprise-kubernetes/master/crds.yaml`

3. Create MongoDB Enterprise Operator with necessary Kubernetes objects:

`kubectl apply -f https://raw.githubusercontent.com/mongodb/mongodb-enterprise-kubernetes/master/mongodb-enterprise.yaml`

## Create a MongoDBOpsManager resource
4. Create the secret which will contain the registration information for the admin user that will be created in Ops Manager by the Operator:

```bash
kubectl create secret generic ops-manager-admin-secret  \
--from-literal=Username="jane.doe@example.com" \
--from-literal=Password="Passw0rd." \
--from-literal=FirstName="Jane" \
--from-literal=LastName="Doe" -n mongodb
```
5. Create the configuration file ops-manager.yaml for the MongoDBOpsManager resource:

```yaml
apiVersion: mongodb.com/v1
kind: MongoDBOpsManager
metadata:
  name: ops-manager
  namespace: mongodb
spec:
  # the version of Ops Manager distro to use
  version: 4.2.4

  # the name of the secret containing admin user credentials.
  adminCredentials: ops-manager-admin-secret

  externalConnectivity:
    type: LoadBalancer

  # the Replica Set backing Ops Manager. 
  # appDB has the SCRAM-SHA authentication mode always enabled
  applicationDatabase:
    members: 3
```
and apply it

```
kubectl apply -f ops-manager.yaml
```
Make sure the Ops Manager resource gets to the “Running” phase. Wait ~5 minutes to start it the first time:

```
kubectl get om -n mongodb
```

6. Open your Ops Manager application in a browser to complete configuration. Locate the URL of the LoadBalancer service which was created by the Operator.

```bash
kubectl get svc ops-manager-svc-ext -o jsonpath='{.status.loadBalancer.ingress[0].hostname}' -n mongodb

A5feeb81aere042bda0fc1bda0a77975-38586691.eu-west-1.elb.amazonaws.com
```
You may need to wait for some time until the DNS gets populated for the Load Balancer address.

7. Open the link `http://<elb-url>:8080` in your browser. Login to Ops Manager using the credentials you specified in the secret created and pass the wizard steps to finish Ops Manager configuration. You’ll reach the admin page.

8. Remove the secret ‘ops-manager-admin-secret’. It won’t be used by the Operator anymore.

```
kubectl delete secret ops-manager-admin-secret -n mongodb
```

## Deploying a MongoDb Replica Set to host your databases and collections

1. Open the Ops Manager application. In the UI, generate a new API key by selecting: “UserName -> Account -> Public API Access”

2. Use this key to create a Secret to store Ops Manager credentials:

```bash
kubectl create secret generic om-jane-doe-credentials  \
--from-literal="user=jane.doe@example.com" \
--from-literal="publicApiKey=<publicKey>"  -n mongodb
```

3. Create a ConfigMap describing the connection to the Ops Manager application. You can use “status.opsmanager.url” to get the value for “baseUrl”:

```bash
kubectl get om ops-manager -o jsonpath='{.status.opsManager.url}'

http://ops-manager-svc.mongodb.svc.cluster.local:8080

kubectl create configmap ops-manager-connection  \
--from-literal="baseUrl=http://ops-manager-svc.mongodb.svc.cluster.local:8080"  -n mongodb
```

4. Create the replica-set.yaml, it will be the yaml to use to be able to deploy any mongodb new replica set.
```yaml
apiVersion: mongodb.com/v1
kind: MongoDB
metadata:
  name: my-replica-set
  namespace: mongodb
spec:
  members: 3
  version: 4.2.2-ent
  type: ReplicaSet

  opsManager:
    configMapRef:
      name: ops-manager-connection
  credentials: om-jane-doe-credentials
```

5. Apply it:

```
kubectl apply -f replica-set.yaml
```

Wait until the ressource enters Running state:

```
kubectl get mdb -n mongodb -w
NAME             TYPE         STATE     VERSION   AGE
my-replica-set   ReplicaSet   Running   4.2.2-ent 12m
```

That's it!

## Credits
MongoDb installation inspired by Anton Lisovenko's [blog post](https://www.mongodb.com/blog/post/running-mongodb-ops-manager-in-kubernetes#:~:text=The%20MongoDB%20Enterprise%20Kubernetes%20Operator%2C%20or%20simply%20the%20Operator%2C%20manages,changing%20these%20settings%20as%20needed).
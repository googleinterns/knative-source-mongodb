module github.com/googleinterns/knative-source-mongodb

go 1.14

require (
	github.com/cloudevents/sdk-go/v2 v2.1.0
	github.com/google/go-cmp v0.4.0
	go.mongodb.org/mongo-driver v1.1.2
	go.uber.org/zap v1.14.1
	k8s.io/api v0.17.6
	k8s.io/apimachinery v0.17.6
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/eventing v0.16.0
	knative.dev/pkg v0.0.0-20200702222342-ea4d6e985ba0
	knative.dev/test-infra v0.0.0-20200630141629-15f40fe97047
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
)

replace k8s.io/code-generator => k8s.io/code-generator v0.17.6

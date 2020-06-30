module github.com/googleinterns/knative-source-mongodb

go 1.14

require (
	go.mongodb.org/mongo-driver v1.1.2
	knative.dev/pkg v0.0.0-20200603222317-b79e4a24ca50
	knative.dev/test-infra v0.0.0-20200606045118-14ebc4a42974
)

replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
)

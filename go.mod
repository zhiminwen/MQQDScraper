module main

go 1.14

replace pkg/k8sDiscovery => ./pkg/k8sDiscovery

replace pkg/k8sExec => ./pkg/k8sExec

require (
	cloud.google.com/go v0.38.0 // indirect
	github.com/Azure/go-autorest/autorest v0.9.0 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/imdario/mergo v0.3.9 // indirect
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v1.6.0
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	k8s.io/api v0.18.3 // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20200603063816-c1c6865ac451 // indirect
	pkg/k8sDiscovery v0.0.0-00010101000000-000000000000
	pkg/k8sExec v0.0.0-00010101000000-000000000000
)

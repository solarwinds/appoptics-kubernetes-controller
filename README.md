
# Kubernetes Controller for AppOptics
  
# Table of Contents  
  
* [What is an AppOptics controller?](#What_is_an_AppOptics_controller?)  
* [Requirements](#Requirements)  
* [Important information](#)  
* [Build, docker build, push...](#Build,-docker-build,-push...)  
* [Deploy appoptics-kubernetes-controller](#Deploy-appoptics-kubernetes-controller)  
* [Run it locally connecting to a k8s cluster](#Run-it-locally-connecting-to-a-k8s-cluster)  
  
  
## What is an AppOptics controller?  
  
The `appoptics-kubernetes-controller` is a Kubernetes controller (a.k.a. an operator) that provides a Kubernetes-native interface for managing select AppOptics resources. Currently, the controller manages the following custom resources:

- `Alerts`, `Dashboards` and `Services`

Using an AppOptics token you provide, the controller will create thes resources your AppOptics account. This controller ensures these AppOptics resources conform to the values you define in the `Spec`.
  
Stated differently, this controller can create/update/delete AppOptics Charts, Services and Alerts.  

## Deployment
###
  
  * A Kubernetes v1.9.0 or greater cluster  
  
## Important information  
  To save a secret containing your AppOptics token to your namespace.
  `make add_token NAMESPACE=<b>Your Namespace</b> TOKEN=<b>APPOPTICS API TOKEN</b>`

  NAMESPACE is optional, if no namespace is passed it will run cluster wide


This will create a secret object with the name "appoptics" in your name space within your CRDs you can reference this by defining your secret name in the spec eg 
```
apiVersion: "appoptics.io/v1"  
kind: "Service"  
metadata:  
  name: exampleservice  
  namespace: Your Namespace  
  finalizers:  
  - appoptics.io  
spec:  
  namespace: "default"  
  <b>secret: "appoptics" </b> 
  data: |-  
    type: "mail"  
    settings:  
      addresses: "support@support.io"  
    title: "SUPPORT"
```
  
## Deploy appoptics-kubernetes-controller   
In the `manifests` folder there are some resources that will help you:  
   
  * `dashboard-crd.yaml` - The Dashboard CRD used by the controller.  
	  * `examples/example-dashboard.yaml` - Just an example of the `dashboard` CRD.  

  * `service-crd.yaml` - The Service CRD used by the controller.  
	  * `examples/example-service.yaml` - Just an example of the `service` CRD.  

  * `alert-crd.yaml` - The Alert CRD used by the controller.  
	  * `examples/example-alert.yaml` - Just an example of the `alert` CRD.  
  
## Run it locally connecting to a k8s cluster  
  
You can run the controller locally! To do this you can just build the controller using `go build` and this will create the binary `appoptics-kubernetes-controller` in your project root path.  
  
To connect the controller to a cluster you need a valid `kubeconfig`, so if you already have your `kubeconfig` in `~/.kube/config` you can run:  
  
```  
NAMESPACE=appoptics-kubernetes-controller RESYNC_SECS=60 ./appoptics-kubernetes-controller --kubeconfig=~/.kube/config -v=1 -logtostderr=true  
```  
  
Note: `-v=1 -logtostderr=true` are not required but it's useful to see some logs.

## Contributing
### Requirements  
  
  * go v1.10.2  
  * [dep](https://github.com/golang/dep)  
  * Kubernetes v1.9.0 or greater cluster

### Building
  
To make things easier, checkout the project to:  
`$GOPATH/src/github.com/solarwinds/appoptics-kubernetes-controller`
  
To build the controller and create a docker container we just need to run `make`. This will:  
  
  * Update project dependencies with `dep ensure`  
  
  * Update k8s auto generated code  
  
  * Run the tests  
  
  * Build a docker image and tag it    
  
After that you can build and push the docker image to `docker.com/solarwinds/appoptics-kubernetes-controller` with:
  
  * To push the image with the tag `canary` and APP_VERSION `canary` you can just run: `make push`  
  
  * If we want to push the image with a different tag (or tags!) you can run: `IMAGE_TAG=belitre make push`, this will push the docker image as `docker.com/solarwinds/appoptics-kubernetes-controller:belitre`
  
  * __IMAGE_TAG__ is used to get the version to tag the resources, the controller will tag resources with the labels:  
  
    * `app=appoptics-kubernetes-controller`  
    * `version=IMAGE_TAG-GIT_COMMIT`  
    
# Questions/Comments

Please open an [issue](/issues). We'd love to hear from you. As a SolarWinds Innovation Project, this adapter is supported in a best-effort fashion.

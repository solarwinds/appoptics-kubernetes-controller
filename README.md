
# appoptics-kubernetes-controller  
  
# Table of Contents  
  
* [What is an AppOptics controller?](#What_is_an_AppOptics_controller?)  
  
* [Requirements](#Requirements)  
  
* [Important information](#)  
  
* [Build, docker build, push...](#Build,-docker-build,-push...)  
  
* [Deploy appoptics-kubernetes-controller](#Deploy-appoptics-kubernetes-controller)  
  
* [Run it locally connecting to a k8s cluster](#Run-it-locally-connecting-to-a-k8s-cluster)  
  
  
## What is an AppOptics controller?  
  
appoptics-kubernetes-controller is a controller (also called some times operator) for a kubernetes cluster.  
  
The controller is designed to watch for a 3 custom resources called `Alerts`, `Dashboards` and `Services`, and using the values from the `Spec` of the resource, it will apply some templates and create this `Dashboard`, `Alerts` or `Service` in the AppOptics account pertaining to the Token secret.  
  
Basically the resources the controller can create/update/delete are Charts, Services, Alerts, and Spaces.  
  
  
## Requirements  
  
  * go v1.10.2  
  
  * [dep](https://github.com/golang/dep)  
  
  * Kubernetes v1.9.0  
  
  
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
  
## Build, docker build, push...  
  
To make things easier, checkout the project to:  
`$GOPATH/src/github.com/appoptics/appoptics-kubernetes-controller`  
  
To build the controller and create a docker container we just need to run `make`. This will:  
  
  * Update project dependencies with `dep ensure`  
  
  * Update k8s auto generated code  
  
  * Run the tests  
  
  * Build a docker image and tag it    
  
After that you can build and push the docker image to `docker.com/appoptics/appoptics-kubernetes-controller` with:  
  
  * To push the image with the tag `canary` and APP_VERSION `canary` you can just run: `make push`  
  
  * If we want to push the image with a different tag (or tags!) you can run: `IMAGE_TAG=belitre make push`, this will push the docker image as `docker.com/appoptics/appoptics-kubernetes-controller:belitre`  
  
  * __IMAGE_TAG__ is used to get the version to tag the resources, the controller will tag resources with the labels:  
  
    * `app=appoptics-kubernetes-controller`  
  
    * `version=IMAGE_TAG-GIT_COMMIT`  
    
  
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
  
Note: `-v=1 -logtostderr=true` are not required but it's useful to see some logs
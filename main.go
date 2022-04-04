/*
Copyright 2017 The Kubernetes Authors.

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
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned"
	aoscheme "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned/scheme"
	informers "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/informers/externalversions"

	"github.com/solarwinds/appoptics-kubernetes-controller/pkg/signals"

	co "github.com/solarwinds/appoptics-kubernetes-controller/pkg/controller"
)

var (
	masterURL  string
	kubeconfig string
)

const controllerAgentName = "appoptics"
const namespaceEnvVar = "NAMESPACE"
const resyncEnvVar = "RESYNC_SECS"

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	aoClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building ao clientset: %s", err.Error())
	}

	resyncInSecs, err := getResync()
	if err != nil {
		klog.Fatalf("Error getting ao resync time: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*time.Duration(resyncInSecs))

	aoInformerFactory, err := getAppOpticsInformerFactory(aoClient, time.Second*time.Duration(resyncInSecs))
	if err != nil {
		klog.Fatalf("Error getting ao namespace: %s", err.Error())
	}
	customScheme := scheme.Scheme
	aoscheme.AddToScheme(customScheme)

	controller := co.NewController(kubeClient, aoClient, aoInformerFactory, controllerAgentName, resyncInSecs)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(stopCh)
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	aoInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func getNamespace() (string, error) {
	return getEnvVar(namespaceEnvVar)
}

func getResync() (int64, error) {
	s, err := getEnvVar(resyncEnvVar)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("Error parsing %s from environment variable %s to integer, error was: %v", s, resyncEnvVar, err)
	}
	return i, nil
}

func getEnvVar(name string) (string, error) {
	v, found := os.LookupEnv(name)
	if !found {
		return "", fmt.Errorf("%s must be set", name)
	}
	if len(v) == 0 {
		return "", fmt.Errorf("%s must not be empty", name)
	}
	return v, nil
}

func getAppOpticsInformerFactory(client *clientset.Clientset, resync time.Duration) (informers.SharedInformerFactory, error) {
	_, namespaced := os.LookupEnv(namespaceEnvVar)
	if namespaced {
		aoNamespace, err := getNamespace()
		if err != nil {
			return nil, err
		}
		return informers.NewFilteredSharedInformerFactory(client, resync, aoNamespace, nil), nil
	}

	return informers.NewSharedInformerFactory(client, resync), nil

}

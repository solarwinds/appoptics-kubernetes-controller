/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned/typed/appopticskubernetescontroller/v1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeAppopticsV1 struct {
	*testing.Fake
}

func (c *FakeAppopticsV1) AppOpticsAlerts(namespace string) v1.AppOpticsAlertInterface {
	return &FakeAppOpticsAlerts{c, namespace}
}

func (c *FakeAppopticsV1) AppOpticsDashboards(namespace string) v1.AppOpticsDashboardInterface {
	return &FakeAppOpticsDashboards{c, namespace}
}

func (c *FakeAppopticsV1) AppOpticsServices(namespace string) v1.AppOpticsServiceInterface {
	return &FakeAppOpticsServices{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeAppopticsV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
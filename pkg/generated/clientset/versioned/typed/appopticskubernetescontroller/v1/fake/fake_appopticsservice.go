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
	"context"

	appopticskubernetescontrollerv1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appopticskubernetescontroller/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAppOpticsServices implements AppOpticsServiceInterface
type FakeAppOpticsServices struct {
	Fake *FakeAppopticsV1
	ns   string
}

var appopticsservicesResource = schema.GroupVersionResource{Group: "appoptics.io", Version: "v1", Resource: "appopticsservices"}

var appopticsservicesKind = schema.GroupVersionKind{Group: "appoptics.io", Version: "v1", Kind: "AppOpticsService"}

// Get takes name of the appOpticsService, and returns the corresponding appOpticsService object, and an error if there is any.
func (c *FakeAppOpticsServices) Get(ctx context.Context, name string, options v1.GetOptions) (result *appopticskubernetescontrollerv1.AppOpticsService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(appopticsservicesResource, c.ns, name), &appopticskubernetescontrollerv1.AppOpticsService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*appopticskubernetescontrollerv1.AppOpticsService), err
}

// List takes label and field selectors, and returns the list of AppOpticsServices that match those selectors.
func (c *FakeAppOpticsServices) List(ctx context.Context, opts v1.ListOptions) (result *appopticskubernetescontrollerv1.AppOpticsServiceList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(appopticsservicesResource, appopticsservicesKind, c.ns, opts), &appopticskubernetescontrollerv1.AppOpticsServiceList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &appopticskubernetescontrollerv1.AppOpticsServiceList{ListMeta: obj.(*appopticskubernetescontrollerv1.AppOpticsServiceList).ListMeta}
	for _, item := range obj.(*appopticskubernetescontrollerv1.AppOpticsServiceList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested appOpticsServices.
func (c *FakeAppOpticsServices) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(appopticsservicesResource, c.ns, opts))

}

// Create takes the representation of a appOpticsService and creates it.  Returns the server's representation of the appOpticsService, and an error, if there is any.
func (c *FakeAppOpticsServices) Create(ctx context.Context, appOpticsService *appopticskubernetescontrollerv1.AppOpticsService, opts v1.CreateOptions) (result *appopticskubernetescontrollerv1.AppOpticsService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(appopticsservicesResource, c.ns, appOpticsService), &appopticskubernetescontrollerv1.AppOpticsService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*appopticskubernetescontrollerv1.AppOpticsService), err
}

// Update takes the representation of a appOpticsService and updates it. Returns the server's representation of the appOpticsService, and an error, if there is any.
func (c *FakeAppOpticsServices) Update(ctx context.Context, appOpticsService *appopticskubernetescontrollerv1.AppOpticsService, opts v1.UpdateOptions) (result *appopticskubernetescontrollerv1.AppOpticsService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(appopticsservicesResource, c.ns, appOpticsService), &appopticskubernetescontrollerv1.AppOpticsService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*appopticskubernetescontrollerv1.AppOpticsService), err
}

// Delete takes name of the appOpticsService and deletes it. Returns an error if one occurs.
func (c *FakeAppOpticsServices) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(appopticsservicesResource, c.ns, name, opts), &appopticskubernetescontrollerv1.AppOpticsService{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAppOpticsServices) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(appopticsservicesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &appopticskubernetescontrollerv1.AppOpticsServiceList{})
	return err
}

// Patch applies the patch and returns the patched appOpticsService.
func (c *FakeAppOpticsServices) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *appopticskubernetescontrollerv1.AppOpticsService, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(appopticsservicesResource, c.ns, name, pt, data, subresources...), &appopticskubernetescontrollerv1.AppOpticsService{})

	if obj == nil {
		return nil, err
	}
	return obj.(*appopticskubernetescontrollerv1.AppOpticsService), err
}

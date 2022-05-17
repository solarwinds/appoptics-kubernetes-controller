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

package v1

import (
	"context"
	"time"

	v1 "github.com/solarwinds/appoptics-kubernetes-controller/pkg/apis/appopticskubernetescontroller/v1"
	scheme "github.com/solarwinds/appoptics-kubernetes-controller/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AppOpticsDashboardsGetter has a method to return a AppOpticsDashboardInterface.
// A group's client should implement this interface.
type AppOpticsDashboardsGetter interface {
	AppOpticsDashboards(namespace string) AppOpticsDashboardInterface
}

// AppOpticsDashboardInterface has methods to work with AppOpticsDashboard resources.
type AppOpticsDashboardInterface interface {
	Create(ctx context.Context, appOpticsDashboard *v1.AppOpticsDashboard, opts metav1.CreateOptions) (*v1.AppOpticsDashboard, error)
	Update(ctx context.Context, appOpticsDashboard *v1.AppOpticsDashboard, opts metav1.UpdateOptions) (*v1.AppOpticsDashboard, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.AppOpticsDashboard, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.AppOpticsDashboardList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AppOpticsDashboard, err error)
	AppOpticsDashboardExpansion
}

// appOpticsDashboards implements AppOpticsDashboardInterface
type appOpticsDashboards struct {
	client rest.Interface
	ns     string
}

// newAppOpticsDashboards returns a AppOpticsDashboards
func newAppOpticsDashboards(c *AppopticsV1Client, namespace string) *appOpticsDashboards {
	return &appOpticsDashboards{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the appOpticsDashboard, and returns the corresponding appOpticsDashboard object, and an error if there is any.
func (c *appOpticsDashboards) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.AppOpticsDashboard, err error) {
	result = &v1.AppOpticsDashboard{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AppOpticsDashboards that match those selectors.
func (c *appOpticsDashboards) List(ctx context.Context, opts metav1.ListOptions) (result *v1.AppOpticsDashboardList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.AppOpticsDashboardList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested appOpticsDashboards.
func (c *appOpticsDashboards) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a appOpticsDashboard and creates it.  Returns the server's representation of the appOpticsDashboard, and an error, if there is any.
func (c *appOpticsDashboards) Create(ctx context.Context, appOpticsDashboard *v1.AppOpticsDashboard, opts metav1.CreateOptions) (result *v1.AppOpticsDashboard, err error) {
	result = &v1.AppOpticsDashboard{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(appOpticsDashboard).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a appOpticsDashboard and updates it. Returns the server's representation of the appOpticsDashboard, and an error, if there is any.
func (c *appOpticsDashboards) Update(ctx context.Context, appOpticsDashboard *v1.AppOpticsDashboard, opts metav1.UpdateOptions) (result *v1.AppOpticsDashboard, err error) {
	result = &v1.AppOpticsDashboard{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		Name(appOpticsDashboard.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(appOpticsDashboard).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the appOpticsDashboard and deletes it. Returns an error if one occurs.
func (c *appOpticsDashboards) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *appOpticsDashboards) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("appopticsdashboards").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched appOpticsDashboard.
func (c *appOpticsDashboards) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.AppOpticsDashboard, err error) {
	result = &v1.AppOpticsDashboard{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("appopticsdashboards").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
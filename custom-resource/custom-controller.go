// A custom controller for MyResource.
// Visit https://github.com/kubernetes/sample-controller for an actual example

// - Kubebuilder: A popular framework for building basic Kubernetes operators using Go.
// - Operator SDK: Another widely used toolkit for developing operators, supporting multiple languages.
// - Client-Go: For lower-level development, provides libraries to interact with the Kubernetes API.

package myresource
var controllerKind = apps.SchemeGroupVersion.WithKind("MyResource")

// Some code

// Watch and sync
func (dc *MyResourceController) Run(workers int, stopCh <-chan struct{})

// Some code

func (dc *MyResourceController) callMyResourceAPI(obj interface{})

// Some code
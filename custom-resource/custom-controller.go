// A custom controller for MyResource.

// Visit https://github.com/kubernetes/sample-controller for an actual example

package myresource
var controllerKind = apps.SchemeGroupVersion.WithKind("MyResource")

// Some code

// Watch and sync
func (dc *MyResourceController) Run(workers int, stopCh <-chan struct{})

// Some code

func (dc *MyResourceController) callMyResourceAPI(obj interface{})

// Some code
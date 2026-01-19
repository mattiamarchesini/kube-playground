module myproject

go 1.21 // Or your Go version

require (
    k8s.io/api v0.30.0
    k8s.io/apimachinery v0.30.0
    k8s.io/client-go v0.30.0
    sigs.k8s.io/controller-runtime v0.18.0
)

// Add require directives for indirect dependencies as needed by go mod tidy
// ...

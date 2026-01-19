package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MyResourceSpec defines the desired state of MyResource
type MyResourceSpec struct {
    // Replicas is the desired number of replicas for the deployment
    // +kubebuilder:validation:Minimum=0
    Replicas *int32 `json:"replicas"`
}

// MyResourceStatus defines the observed state of MyResource
type MyResourceStatus struct {
    // Phase represents the current phase of the resource (e.g., Processing, Available, Failed)
    Phase string `json:"phase,omitempty"`
    // ObservedReplicas is the actual number of replicas observed for the deployment
    ObservedReplicas int32 `json:"observedReplicas,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MyResource is the Schema for the myresources API
type MyResource struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   MyResourceSpec   `json:"spec,omitempty"`
    Status MyResourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MyResourceList contains a list of MyResource
type MyResourceList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []MyResource `json:"items"`
}

// init registers the types with the SchemeBuilder
func init() {
    SchemeBuilder.Register(&MyResource{}, &MyResourceList{})
}

package controllers

import (
    "context"
    "fmt"
    "time"

    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/types"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"

    // Import your API definition
    mygroupv1alpha1 "myproject/api/v1alpha1"
)

// MyResourceReconciler reconciles a MyResource object
type MyResourceReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mygroup.example.com,resources=myresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mygroup.example.com,resources=myresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mygroup.example.com,resources=myresources/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the main control loop logic
func (r *MyResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    logger.Info("Reconciling MyResource")

    // 1. Fetch the MyResource instance
    myResource := &mygroupv1alpha1.MyResource{}
    if err := r.Get(ctx, req.NamespacedName, myResource); err != nil {
        if apierrors.IsNotFound(err) {
            // Request object not found, could have been deleted after reconcile request.
            // Return and don't requeue
            logger.Info("MyResource resource not found. Ignoring since object must be deleted")
            return ctrl.Result{}, nil
        }
        // Error reading the object - requeue the request.
        logger.Error(err, "Failed to get MyResource")
        return ctrl.Result{}, err
    }

    // Update status phase to Processing
    if myResource.Status.Phase != "Processing" {
        myResource.Status.Phase = "Processing"
        if err := r.Status().Update(ctx, myResource); err != nil {
            logger.Error(err, "Failed to update MyResource status to Processing")
            return ctrl.Result{}, err
        }
        // Requeue immediately after status update
        return ctrl.Result{Requeue: true}, nil
    }


    // 2. Check if the Deployment already exists, if not create a new one
    foundDeployment := &appsv1.Deployment{}
    err := r.Get(ctx, types.NamespacedName{Name: myResource.Name, Namespace: myResource.Namespace}, foundDeployment)
    if err != nil && apierrors.IsNotFound(err) {
        // Define a new deployment
        dep := r.deploymentForMyResource(myResource)
        logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
        if err = r.Create(ctx, dep); err != nil {
            logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
            myResource.Status.Phase = "Failed"
            _ = r.Status().Update(ctx, myResource) // Best effort status update
            return ctrl.Result{}, err
        }
        // Deployment created successfully - return and requeue to check status later
        return ctrl.Result{RequeueAfter: time.Minute}, nil
    } else if err != nil {
        logger.Error(err, "Failed to get Deployment")
        myResource.Status.Phase = "Failed"
        _ = r.Status().Update(ctx, myResource) // Best effort status update
        return ctrl.Result{}, err
    }

    // 3. Ensure the deployment size is the same as the spec
    size := myResource.Spec.Replicas
    if size == nil {
        // Default replicas to 1 if not specified
        defaultReplicas := int32(1)
        size = &defaultReplicas
    }

    if *foundDeployment.Spec.Replicas != *size {
        logger.Info("Deployment replicas size mismatch", "Expected", *size, "Found", *foundDeployment.Spec.Replicas)
        foundDeployment.Spec.Replicas = size
        if err = r.Update(ctx, foundDeployment); err != nil {
            logger.Error(err, "Failed to update Deployment", "Deployment.Namespace", foundDeployment.Namespace, "Deployment.Name", foundDeployment.Name)
            myResource.Status.Phase = "Failed"
            _ = r.Status().Update(ctx, myResource) // Best effort status update
            return ctrl.Result{}, err
        }
        // Spec updated - return and requeue
        return ctrl.Result{RequeueAfter: time.Minute}, nil
    }

    // 4. Update the MyResource status with the observed replicas
    observedReplicas := foundDeployment.Status.ReadyReplicas
    if myResource.Status.Phase != "Available" || myResource.Status.ObservedReplicas != observedReplicas {
        logger.Info("Updating MyResource status", "Observed Replicas", observedReplicas)
        myResource.Status.Phase = "Available"
        myResource.Status.ObservedReplicas = observedReplicas
        if err := r.Status().Update(ctx, myResource); err != nil {
            logger.Error(err, "Failed to update MyResource status")
            // Don't return error, maybe just transient, requeue
            return ctrl.Result{RequeueAfter: time.Minute}, nil
        }
    }


    logger.Info("Reconciliation finished")
    // If everything is reconciled, stop processing
    return ctrl.Result{}, nil
}

// deploymentForMyResource returns a Deployment object for the MyResource
func (r *MyResourceReconciler) deploymentForMyResource(m *mygroupv1alpha1.MyResource) *appsv1.Deployment {
    ls := labelsForMyResource(m.Name)
    replicas := m.Spec.Replicas
    if replicas == nil {
        defaultReplicas := int32(1)
        replicas = &defaultReplicas
    }


    dep := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      m.Name,
            Namespace: m.Namespace,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: ls,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: ls,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Image: "nginx:latest", // Example image
                        Name:  "webserver",
                        Ports: []corev1.ContainerPort{{
                            ContainerPort: 80,
                            Name:          "http",
                        }},
                    }},
                },
            },
        },
    }
    // Set MyResource instance as the owner and controller
    ctrl.SetControllerReference(m, dep, r.Scheme)
    return dep
}

// labelsForMyResource returns the labels for selecting the resources
// belonging to the given MyResource CR name.
func labelsForMyResource(name string) map[string]string {
    return map[string]string{"app": "myresource", "myresource_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&mygroupv1alpha1.MyResource{}).
        Owns(&appsv1.Deployment{}). // Watch Deployments owned by MyResource
        Complete(r)
}

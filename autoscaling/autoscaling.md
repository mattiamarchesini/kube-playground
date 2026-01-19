# Autoscaling

3 types of auto-scalers

- cluster (not part of CKA exam)
- horizontal pod
- vertical pod

## Horizontal pod scaling

In this context, it means increasing/decreasing the number of pods. It's the recommended option for micro/stateless
services.

To do it manually, one would have to monitor resources using a command like

```bash
kubectl top pod my-app-pod
```

Then, when the resource is near the threshold, run

```bash
kubectl scale deployment my-app --replicas=3
```

The horizontal auto-scaler:

- tracks multiple types of metrics
- adds pods
- balances thresholds

You can configure the horizontal auto-scaler by running a command like this

```bash
kubectl autoscale deployment my-app --cpu-percent=50 --min=1 --max=10
```

Get the status of the auto-scaler by running

```bash
kubectl get hpa
```

See `hpa.yaml` for a declarative approach to the horizontal auto-scaler.

## Vertical pod scaling

In this context, it means increasing/decreasing CPU and memory resources of existing pods. This requires a restart of
the pods to apply the new values. Vertical scaling is recommended for stateful workloads and CPU/memory-heavy apps.

Unlike HPA; the Vertical Pod Autoscaler doesn't come built-in, so we must deploy it as a pod:

```bash
kubectl apply -f https://github.com/kubernetes/autoscaler/releases/latest/download/vertical-pod-autoscaler.yaml
```

The VPA is divided into 3 different pods:

- **recommender**: responsible for monitoring resources, collect usage data and provide recommendations for CPU and
  memory values
- **updater**: detects pods and evicts them when they are running with sub-optimal resources and an update is needed
- **admission-controller**: intervenes in the pod creation process, using the recommendations from the recommender to
  update the pod specs accordingly.

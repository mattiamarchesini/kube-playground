# Metrics

For Minikube:

```sh
minikube addons enable metrics-server
```

For others

```sh
git clone https://github.com/kubernetes-incubator/metrics-server.git

# Deploy pods, services and roles needed for performance metrics
kubectl create -f metrics-server/deploy/1.8+/
```

After some time it will start collecting data. Watch them using

```sh
kubectl top [pod, node]
kubectl logs <pod name>
```

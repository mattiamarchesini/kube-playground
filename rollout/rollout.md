# Rolling updates and Rollbacks

See the status

```sh
kubectl rollout status deployment/myapp-depl
```

See revisions and history of rollouts

```sh
kubectl rollout history deployment/myapp-depl
```

Go back

```sh
kubectl rollout undo deployment/myapp-depl
```

Default deployment strategy (RollingUpdate) is to not cause any downtime.

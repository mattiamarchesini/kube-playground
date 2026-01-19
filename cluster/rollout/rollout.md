# Rolling updates and Rollbacks

See the status

```sh
kubectl rollout status deployment/myapp-depl
```

See revisions and history of rollouts

```sh
kubectl rollout history deployment/myapp-depl
```

Update to a new image version:

```bash
kubectl set image deployment/myapp-depl nginx=nginx:1.17
```

Go back to previous version:

```sh
kubectl rollout undo deployment/myapp-depl
```

Default deployment strategy (RollingUpdate) is to not cause any downtime.

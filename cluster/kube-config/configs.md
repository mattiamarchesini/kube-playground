# Kube config

You can specify a configuration file for `kubectl` using the `--kubeconfig` flag or by setting the `KUBECONFIG`
variable.

Inside a `kubeconfig` file, a context defines which user will have access to which cluster. To use a context different
from the one specified in `current-context`, run

```bash
kubectl config use-context user01@another-context
```

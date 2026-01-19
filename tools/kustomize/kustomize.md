# Kustomize

Lets you apply patches to your yaml definition files so you can have different customizations for different
environments.

To output the build of the customization (files are inside the `k8s` folder):

```bash
kustomize build k8s/
```

To apply the build:

```bash
kustomize build k8s/ | kubectl apply -f -

# Or
kubectl apply -k k8s/

# Or
kubectl kustomize build k8s/ | kubectl apply -f -
```

To delete use `kubectl delete` instead of `kubectl apply`.

## Transformers

They can be built-in or custom made. They allow to make configuration changes across all the resources.

Some examples of built-in transformers you can add to `kustomize.yaml`:

- `commonLabels`: add common labels to all the resources
- `commonAnnotations`: add common annotations to the resources
- `namePrefix/Suffix`: add a common prefix/suffix to the resources
- `namespace`: add a common namespace
- `images`: change the images utilized by pods. Need to specify a list of objects with `name` and other properties like
  `newName` or `newTag`

## Patches

You can also apply patches. They provide a more targeted approach compared to transformers.

Here's how a patch inside `kustomization.yaml` would look like:

```yaml
patches:
  - target:
      kind: Deployment
      name: api-depl
    patch: |-
      - op: replace # add, remove
        path: /spec/replicas
        value: 5
```

This type of patch follows the [**JSON Patch RFC**](https://datatracker.ietf.org/doc/html/rfc6902).

You can also do a **Strategic merge Patch**, where it would figure out what's changed and merge the old and new config
together:

```yaml
patches:
  - patch: |-
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: api-depl
      spec:
        replicas: 5
```

To delete a property with this type of patch, just set it to `null`.

For both patch methods, you could also just specify the path of a yaml file containing all the updated values:

```yaml
patches:
  - path: path-to-patch-file.yaml
```

To patch an element of a list:

```yaml
# patch a list using json patch
patches:
  - target:
      kind: Deployment
      name: api-depl
    patch: |-
      - op: replace
        path: /spec/template/spec/containers/0  # patch element 0 of the list spec.template.spec.containers
        value:
          image: a-different-nginx
      - op: add
        path: /spec/template/spec/containers/-  # add a new element to the list
        value:
          name: another-nginx
          image: another-nginx
---
# delete an element from a list using strategic merge
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-depl
spec:
  template:
    spec:
      containers:
        - $patch: delete
          name: database
```

## Overlays

Overlays let you customize on a environment basis. You just add a base configuration then modify it with patches for
each environment. Remember to reference the base like this:

```yaml
resources:
  - ../../base
# in older versions
# bases:
#   - ../../base
```

## Components

Components can be reused and included in multiple overlays. They are useful for group of features that should be enabled
only for some environments.

The `components` field in `kustomization.yaml` is actually considered deprecated. While it still works in current
versions of Kustomize, it is no longer the recommended approach for composing configurations, and you may see warnings
when you use it.

The Recommended "New Way" is you simply treat your "component" as another base. Then, in your final overlay, you include
both. So the new Directory Structure will be:

```text
my-app/
├── base/
│   ├── deployment.yaml
│   └── kustomization.yaml
├── monitoring-base-components/  # Formerly the "component"
│   ├── service-monitor.yaml
│   └── kustomization.yaml
└── overlays/
    └── production/
        └── kustomization.yaml # This file combines everything
```

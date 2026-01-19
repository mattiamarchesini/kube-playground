# Cluster management tools

## Kubecm

kubecm is a configuration file management tool. It was created to solve the problem of those who
find themselves with dozens of separate `kubeconfig.yaml` files, one for each cluster.

Base commands:

- `add`: Merge one or more new kubeconfig files into your main file `~/.kube/config`.
- `delete <context-name>`: Removes a context (and the associated cluster/user) from your configuration file.
- `rename <old> <new>`: Rename a context cleanly.
- `switch <another-context>`: Switch to a context

## Kubectx and Kubens

**Kubectx** lets you quickly switch context between clusters in a multi-cluster environment.

**Installation**:

```bash
sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
sudo ln -s /opt/kubectx/kubectx /usr/local/bin/kubectx
```

**Syntax**:

To list all contexts:

```bash
kubectx
```

To switch to a new context:

```bash
kubectx <context_name>
```

To switch back to previous context:

```bash
kubectx -
```

To see current context:

```bash
kubectx -c
```

**Kubens** allows users to switch between namespaces quickly with a simple command.

**Installation**:

```bash
sudo git clone https://github.com/ahmetb/kubectx /opt/kubectx
sudo ln -s /opt/kubectx/kubens /usr/local/bin/kubens
```

**Syntax**:

To switch to a new namespace:

```bash
kubens <new_namespace>
```

To switch back to previous namespace:

```bash
kubens -
```

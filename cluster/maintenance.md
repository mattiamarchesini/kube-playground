# Cluster maintenance

Every release of kubernetes is composed of these components that all share the same versioning number:

- kube-apiserver (the main component)
- controller-manager
- kube-scheduler
- kubelet
- kube-proxy
- kubectl

Other fundamental components, like etcd and CoreDNS, are separate projects that follow their own versioning.

## Install

First, install and configure a container runtime. We will use `containerd` and assume the host is running Ubuntu server:

```bash
sudo apt install containerd
sudo modprobe overlay
sudo modprobe br_netfilter

# Make loading of modules permanent
sudo tee /etc/modules-load.d/containerd.conf <<EOF
overlay
br_netfilter
EOF

# Configure System Settings for Networking to handle bridged network traffic for pods.
sudo tee /etc/sysctl.d/99-kubernetes-cri.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF

sudo sysctl --system

# Configure containerd
sudo mkdir -p /etc/containerd
sudo containerd config default | sudo tee /etc/containerd/config.toml
# Set the cgroup driver to systemd
sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml

# It would be better to adjust [plugins."io.containerd.grpc.v1.cri"].sandbox_image
# to the exact version used by the kubelet

sudo systemctl enable --now containerd
```

Then, install kubernetes using `kubeadm`:

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gpg

# Add the official Kubernetes keyrings and apt repository
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.34/deb/Release.key | sudo gpg \
                    --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.34/deb/ /' | \
                    sudo tee /etc/apt/sources.list.d/kubernetes.list

# Then install the k8s components
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl

# Set the components to not upgrade automatically
sudo apt-mark hold kubelet kubeadm kubectl

# By default, kubelet won't work when swap is enabled. Disable it and Remove it from fstab as well.
sudo swapoff -a
sudo sed -i.bak '/ swap /d' /etc/fstab

sudo systemctl enable --now kubelet.service
```

On the control-plane, run:

```bash
# Create config and give you the token for the worker nodes to join
sudo kubeadm init

# Copy the admin config file to your user's config directory to be able to run kubectl as normal user
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# You need to install a CNI plugin.
# We'll use Calico (with Tigera operator), which is a popular and powerful choice.
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/tigera-operator.yaml
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/custom-resources.yaml

# Need to restart containerd or it won't pick up the changes made by calico in /etc/cni/net.d/
sudo systemctl restart containerd
```

On the worker node, run:

```bash
sudo kubeadm join <CONTROL_PLANE_IP>:<PORT> --token <TOKEN> --discovery-token-ca-cert-hash <HASH>
```

Done! Now you should be able to see the nodes on the control plane using

```bash
kubectl get nodes
```

### Install Addons

Add-ons extend the functionality of Kubernetes. Visit
https://kubernetes.io/docs/concepts/cluster-administration/addons/

### Install a Network Plugin

You can find details about the network plugins in the following documentation:

- https://kubernetes.io/docs/concepts/cluster-administration/addons/#networking-and-network-policy
- https://kubernetes.io/docs/concepts/cluster-administration/networking/#how-to-implement-the-kubernetes-networking-model

There are several plugins available and these are some. If there are multiple CNI configuration files in the directory,
the kubelet uses the configuration file that comes first by name in lexicographic order.

- **Calico**: Is said to be the most advanced CNI network plugin. It's recommended to also install the Tigera operator
  first

  ```bash
  kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/tigera-operator.yaml
  kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.30.3/manifests/custom-resources.yaml
  ```

- **Flannel**: Simple and easy layer 3 network fabric configuration. As of now flannel does not support kubernetes
  network policies.

  ```bash
  kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/2140ac876ef134e0ed5af15c65e414cf26827915/Documentation/kube-flannel.yml
  ```

- **Weave Net**: To install

  ```bash
  kubectl apply -f https://github.com/weaveworks/weave/releases/download/v2.8.1/weave-daemonset-k8s.yaml
  ```

## Update

When managing your own cluster, it's fine if the controller-manager and the kube-scheduler are one version _behind_ the
kube-apiserver; while the kubelet and kube-proxy component can be as far as 2 versions behind. The kubectl cli can be
one version higher or lower than kube-apiserver.

The recommended approach is to upgrade one _minor_ version at a time. There is a correct order of commands to upgrade
the cluster without any downtime.
Use any text editor you prefer to open the file that defines the Kubernetes apt repository.

vim /etc/apt/sources.list.d/kubernetes.list

Update the version in the URL to the next available minor release. Let's say v1.33.

```text
deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.33/deb/ /
```

After making changes, proceed with the following

```bash
kubectl drain controlplane --ignore-daemonsets
apt update
apt-cache madison kubeadm
```

Based on the version information displayed by `apt-cache madison`, it indicates that for Kubernetes version 1.33.0, the
available package version is 1.33.0-1.1. Therefore, to install kubeadm for Kubernetes v1.33.0:

```bash
apt-get install kubeadm=1.33.0-1.1
```

Run the following command to upgrade the Kubernetes cluster:

```bash
kubeadm upgrade plan v1.33.0
kubeadm upgrade apply v1.33.0
```

Now, upgrade the version and restart Kubelet. Also, mark the node (in this case, the "controlplane" node) as
schedulable:

```bash
apt-get install kubelet=1.33.0-1.1
systemctl daemon-reload
systemctl restart kubelet
kubectl uncordon controlplane
```

Before draining node01, if the controlplane gets taint during an upgrade, we have to remove it.

Identify the taint first:

```bash
kubectl describe node controlplane | grep -i taint
```

Remove the taint with help of "kubectl taint" command:

```bash
kubectl taint node controlplane node-role.kubernetes.io/control-plane:NoSchedule-
```

Verify it, the taint has been removed successfully:

```bash
kubectl describe node controlplane | grep -i taint
```

Now, drain the node01:

```bash
kubectl drain node01 --ignore-daemonsets
```

SSH into node01 and modify `/etc/apt/sources.list.d/kubernetes.list` as seen for the control plane. After making
changes, proceed with the update:

```bash
apt update
apt-cache madison kubeadm # get latest minor version info

apt-get install kubeadm=1.33.0-1.1
```

Then Upgrade the worker node using `kubeadm`:

```bash
kubeadm upgrade node
```

Now, upgrade the version and restart Kubelet:

```bash
apt-get install kubelet=1.33.0-1.1
systemctl daemon-reload
systemctl restart kubelet
```

At last, Back on the controlplane node:

```bash
kubectl uncordon node01
kubectl get pods -o wide | grep gold # make sure this is scheduled on a node
```

## Filter objects info with kubectl and JSONPath

You can use [JSONPath](https://en.wikipedia.org/wiki/JSONPath) to filter informations. Example:

```bash
kubectl get pods -o=jsonpath='{.items[0].spec.containers[0].image}'
```

You can have custom output and more than one query:

```bash
kubectl get nodes -o=jsonpath='Name: {.items[*].metadata.name} {"\n"}CPU: {.items[*].status.capacity.cpu}'
```

Or custom columns:

```bash
kubectl get nodes -o=custom-columns='NODE:.metadata.name,CPU:.status.capacity.cpu'
```

Or sort them by a specific field:

```bash
kubectl get nodes --sort-by=.metadata.name
```

## Make a backup of ETCD

```bash
ETCDCTL_API=3 etcdctl snapshot save /opt/etcd-backup.db \
  --cacert /etc/kubernetes/pki/etcd/ca.crt \
  --cert /etc/kubernetes/pki/etcd/server.crt \
  --key /etc/kubernetes/pki/etcd/server.key
```

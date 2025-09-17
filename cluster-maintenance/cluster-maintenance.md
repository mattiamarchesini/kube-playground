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

First, install and configure a container runtime (`containerd`)

```console
sudo apt installl containerd
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

Then, install kubernetes the `kubeadm` way:

```console
# Install dependencies
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gpg

# Add the official Kubernetes keyrings and apt repository
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.34/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.34/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# Then install the k8s components
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl

# Set the components to not upgrade automatically
sudo apt-mark hold kubelet kubeadm kubectl

# Enable the kubelet service
sudo swapoff -a # Disable swap or kubelet won't work (remove it from /etc/fstab as well)
sudo systemctl enable --now kubelet
```

On the control-plane, run:

```console
# Will create the needed configuration and
# also give you the token for the worker node to join
sudo kubeadm init

# Copy the admin config file to your user's config directory
# to be able to run kubectl as normal user
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config

# You need to install a CNI plugin. We'll use Calico,
# which is a popular and powerful choice.
kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/refs/heads/master/manifests/calico.yaml
# Need to restart containerd or it won't pick up the changes in /etc/cni/net.d/ made by calico
sudo systemctl restart containerd
```

On the worker node, run:

```console
sudo kubeadm join <CONTROL_PLANE_IP>:<PORT> --token <TOKEN> --discovery-token-ca-cert-hash <HASH>
```

Done! Now you should be able to see the nodes on the control plane using

```console
kubectl get nodes
```

## Update

When managing your own cluster, it's fine if the controller-manager and the kube-scheduler are one version _behind_ the
kube-apiserver; while the kubelet and kube-proxy component can be as far as 2 versions behind. The kubectl cli can be
one version higher or lower than kube-apiserver.

The recommended approach is to upgrade one _minor_ version at a time. There is a correct order of commands to upgrade
the cluster without any downtime.

On the control-plane node:

```console
# Gives you info about latest versions
kubeadm upgrade plan

# First upgrade kubeadm and kubectl using the package manager
sudo apt upgrade -y kubeadm=1.12.0-00 kubectl=1.12.0-00

# Then the other cluster components using kubeadm
kubeadm upgrade apply 1.12.0-00

# And kubelet at last using the package manager
sudo apt upgrade -y kubelet=1.12.0-00
sudo systemctl daemon-reload # Reloads all systemd unit files, just in case kubelet's one was modified
sudo systemctl restart kubelet
```

Then on the worker nodes (one at a time):

```console
# Move the worload to the other nodes. Also cordons the node as unschedulable while upgrading.
kubectl drain node-1

sudo apt upgrade -y kubeadm=1.12.0-00 kubelet=1.12.0-00

# Needs to be after updating kubelet
kubeadm upgrade node
```

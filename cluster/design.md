# Designing a cluster

When designing a cluster, you should consider these factors:

- Purpose:
  - Education: use minikube or a single-node cluster with kubeadm
  - Development: one control plane and multiple workers with kubeadm/GCP/AWS/Azure
  - Production: high-availability, multiple control planes and workers. Managed cloud solutions are highly recommended
- Cloud or On-Prem
- Workload
  - how many?
  - what kind? (web, big-data, analytics...)
  - resource requirements: CPU intensive vs Memory intensive
  - traffic: heavy or burst

Here's a table with recommended machine instance types for AWS and GCP

| Number of nodes | Use case                  | AWS recommendation               | GCP recommendation                  |
| :-------------- | :------------------------ | :------------------------------- | :---------------------------------- |
| 1 - 5           | Development/Testing       | `t3.large` (2 vCPU / 8 GiB)      | `e2-standard-2` (2 vCPU / 8 GiB)    |
| 6 - 20          | Small Production Loads    | `m6g.xlarge` (4 vCPU / 16 GiB)   | `n2-standard-4` (4 vCPU / 16 GiB)   |
| 21 - 100        | Standard Production Loads | `m6g.2xlarge` (8 vCPU / 32 GiB)  | `n2-standard-8` (8 vCPU / 32 GiB)   |
| 100+            | Large Production Loads    | `m6g.4xlarge` (16 vCPU / 64 GiB) | `n2-standard-16` (16 vCPU / 64 GiB) |

An individual control-plane node often has higher baseline requirements than an individual worker node.

In large clusters, you might want to separate etcd from the control plane and give it its own cluster.

## Choosing the right infrastructure

When deploying a cluster, consider using one of these solutions:

- OpenShift: It provides tools and a UI to manage k8s constructs. It easily integrates with CI/CD pipelines.
- Cloud Foundry container runtime and its cli tool `bosh` are a popular FOSS choice
- VMWare Cloud PKS: use it to leverage your VMWare environment for Kubernetes
- Vagrant: provides scripts to deploy clusters on different cloud providers

For deploying a cluster for a hosted solution, you can try:

- Google Kubernetes Engine (GKE)
- Azure Kubernetes Service
- Amazon Elastic Container Service for Kubernetes (EKS)
- OpenShift online: RedHat offers a managed version of OpenShift

## High Availability

It means having redundancy for every component in the cluster (especially the control plane) so you can handle updates
smoothly and not have a single point of failure.

A load balancer should be set in place so requests to the API server don't get replicated and are instead evenly
splitted between the master nodes. `kubectl` should then be configured to point directly to the load balancer.

For some processes like `kube-scheduler` and `kube-controller-manager`, an election is held to decide which node should
have a lease and be active while the others remain in a hot-standby state. By default, every 2 seconds (`retry-period`)
the leader tries to renew its lease and non-leaders check for it. If the lease is not renewed before 10 seconds
(`renew-deadline`) then it becomes invalid and all the other nodes race to become the leader. Whatever happens, the
lease is automatically invalid after 15 seconds (`lease-duration`) to prevent a "split-brain" scenario and ensure an old
leader will terminate itself even if it's network-isolated.

In a HA scenario, you should run etcd clusters on dedicated machines separated from the master nodes to guarantee
resource requirements. In a etcd cluster, data is replicated without race conditions using the
[Raft Consensus Algorithm ](https://raft.github.io/):

- before the leader is selected, a random timer is kicked off on every node. The first to finish its timer notifies the
  others and tries to become the leader
- the other nodes respond with their vote, electing the leader
- the leader will periodically send an heartbeat to the other nodes, informing them that he is still the leader.
- whichever node gets a data update will first sends it to the leader, who will then make sure the other nodes get the
  same data
- If the majority ($\lfloor n/2 \rfloor +1$) of the nodes received the update, then the "write" is considered valid.
  - Because of this, having 2 instances is the same as having 1, since the "quorum" can't be reached if one of them
    fails. So 3 (fault tolerance of 1 node) is the minimum number of nodes recommended and, overall, is recommended to
    have an odd number of nodes.
  - An even number may also leave the cluster with no quorum if the network is splitted exactly in 2.
  - 5 nodes give you a fault tolerance of 2 nodes, which is enough

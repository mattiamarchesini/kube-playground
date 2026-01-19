# Troubleshooting

## Logs

You can use `kubectl logs` to view logs about different resources.

Also log locations to check:

- /var/log/pods
- /var/log/containers
- crictl ps + crictl logs
- docker ps + docker logs (if Docker is used)
- kubelet logs: /var/log/syslog or journalctl -u kubelet.service

## Application failure

https://kubernetes.io/docs/tasks/debug/debug-application/

**Checking accessibility**: make a map of every object and check every link until you find the root issue. If you are
deploying a web server, you can try using `curl` to check.

If something is not right, first check the service

```bash
kubectl describe service myservice
```

Check if `Selector` and `Endpoints` match.

Next, check the pod itself and its logs to be sure it's in a running state.

```bash
kubectl get pod
kubectl describe pod mypod
kubectl logs mypod
```

When watching logs, you can use the `-f` option to follow the stream or the `--previous` option to view the log of a
previous pod that already failed

## Nodes failure

First check the nodes status

```bash
kubectl get nodes
```

Look for flags like OutOfDisk, MemoryPressure or DiskPressure that indicate problems with resources for the node

```bash
kubectl describe node worker-1
```

If the flags are set to "Unknown", then the node is probably not reaching the cluster. Check the last heartbeat of the
node in the description.

Check the node itself for cpu/memory/disk. You can use commands like `top` and `df`.

Check the kubelet service status:

```bash
service kubelet status
sudo journalctl -u kubelet
```

Check `/var/lib/kubelet/config.yaml` and `/etc/kubernetes/kubelet.conf` for misconfiguration.

Check the certificates for expiration or belonging to the wrong group

```bash
openssl x509 -in /var/lib/kubelet/worker-1.crt -text
```

## Controlplane nodes failure

Check nodes and pods status

```bash
kubectl get nodes
kubectl get pods
```

Check control plane pods

```bash
kubectl get pods -n kube-system
```

Also check the manifests for the control plane pods found in `/etc/kubernetes/manifests`

Check control plane services

```bash
service kube-apiserver status
service kube-controller-manager status
service kube-scheduler status
service kube-proxy status
service kubelet status
```

View the logs of the api server

```bash
kubectl logs kube-apiserver-master -n kube-system
sudo journalctl -u kube-apiserver
```

## Network failure

References:

- Debug Service issues:
  https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/
- DNS Troubleshooting:
  https://kubernetes.io/docs/tasks/administer-cluster/dns-debugging-resolution/

### CoreDNS

CoreDNS's memory usage is predominantly affected by the number of Pods and Services in the cluster. Other factors
include the size of the filled DNS answer cache, and the rate of queries received (QPS) per CoreDNS instance.

1. CoreDNS pods in pending state: first check network plugin is installed.
2. CoreDNS pods in CrashLoopBackOff or Error state:

   If you have nodes that are running SELinux with an older version of Docker you might experience a scenario where the
   CoreDNS pods are not starting. To solve that you can try one of the following options:

   - a) Upgrade to a newer version of Docker.
   - b) Disable SELinux.
   - c) Modify the CoreDNS deployment to set allowPrivilegeEscalation to true:
     ```bash
     kubectl -n kube-system get deployment coredns -o yaml | \
     sed 's/allowPrivilegeEscalation: false/allowPrivilegeEscalation: true/g' | \
     kubectl apply -f -
     ```
   - d) Another cause for CoreDNS to have CrashLoopBackOff is when a CoreDNS Pod deployed in Kubernetes detects a loop.

     There are many ways to work around this issue, some are listed here:

     - Add the following to your kubelet yaml config: resolvConf: <path-to-your-real-resolv-conf-file> This flag tells
       kubelet to pass an alternate resolv.conf to Pods. For systems using systemd-resolved,
       /run/systemd/resolve/resolv.conf is typically the location of the "real" resolv.conf, although this can be
       different depending on your distribution.
     - Disable the local DNS cache on host nodes, and restore /etc/resolv.conf to the original.
     - A quick fix is to edit your Corefile, replacing forward . /etc/resolv.conf with the IP address of your upstream
       DNS, for example forward . 8.8.8.8. But this only fixes the issue for CoreDNS, kubelet will continue to forward
       the invalid resolv.conf to all default dnsPolicy Pods, leaving them unable to resolve DNS.

3. CoreDNS pods and kube-dns service are working fine: check the kube-dns service has valid endpoints.
   ```bash
   kubectl -n kube-system get ep kube-dns
   ```
   If there are no endpoints for the service, inspect the service and make sure it uses the correct selectors and ports.

### Kube-Proxy

If you run `kubectl describe ds kube-proxy -n kube-system` you can see that the kube-proxy binary runs with following command inside the kube-proxy container.

```output
Command:
      /usr/local/bin/kube-proxy
      --config=/var/lib/kube-proxy/config.conf
      --hostname-override=$(NODE_NAME)
```

The config file `/var/lib/kube-proxy/config.conf` defines the clusterCIDR, kubeproxy mode, ipvs, iptables, bindaddress, kube-config etc.

Troubleshooting issues related to kube-proxy

1. Check kube-proxy pod in the kube-system namespace is running.
2. Check kube-proxy logs.
3. Check configmap is correctly defined and the config file for running kube-proxy binary is correct.
4. kube-config is defined in the config map.
5. check kube-proxy is running inside the container

   ```output
   $ sudo netstat -plan | grep kube-proxy

    tcp 0 0 0.0.0.0:30081 0.0.0.0:_ LISTEN 1/kube-proxy
    tcp 0 0 127.0.0.1:10249 0.0.0.0:_ LISTEN 1/kube-proxy
    tcp 0 0 172.17.0.12:33706 172.17.0.12:6443 ESTABLISHED 1/kube-proxy
    tcp6 0 0 :::10256 :::\* LISTEN 1/kube-proxy
   ```

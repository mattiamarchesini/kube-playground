# TLS Certificates

Most core k8s components need a TLS certificate and key to securely interact with each other. Because of this, a cluster
needs at least one Certificate Authority, which will need its own certificate and key:

```bash
# Generate key
openssl genrsa -out ca.key 2048

# Generate certificate signing request
openssl req -new -key ca.key -subj "/CN=KUBERNETES-CA" -out ca.csr

# Generate signed certificate (self-signed)
openssl x509 -req -in ca.csr -signkey ca.key -out ca.crt
```

Then you need certificate and key for an admin user:

```bash
# Generate key
openssl genrsa -out admin.key 2048

# Generate certificate signing request (add admin user to the system:masters group)
openssl req -new -key admin.key -subj "/CN=kube-admin/OU=system:masters" -out admin.csr

# Generate signed certificate (SIGNED BY K8S CERTIFICATE AUTHORITY!)
openssl x509 -req -in admin.csr -CA ca.crt -CAkey ca.key -out admin.crt
```

You need to generate the other client certificates for clients: kube-scheduler, kube-controller-manager, kube-proxy.

The etcd server needs a server certificate (`-subj "/CN=etcd-server"`), while The kube-apiserver and kubelet-server need
both.

If you deploy more copies of etcd (high availability) you will need peer certificates for them. For the api servers you
will need a `openssl.cnf` config file:

```toml
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation,
subjectAltName = @alt_names
[alt_names] # Various names that refer to the api-server
DNS.1 = kubernetes
DNS.2 = kubernetes.default
DNS.3 = kubernetes.default.svc
DNS.4 = kubernetes.default.svc.cluster.local
IP.1 = 10.96.0.1 # IPs of api-servers for each control-plane nodes
IP.2 = 172.17.0.87
```

/etc/kubernetes/pki

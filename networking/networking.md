# Networking

By default, the Kubernetes API server listens on port 6443 on the first non-localhost network interface, protected by
TLS. In a typical production Kubernetes cluster, the API serves on port 443.

## Ingress and services

Here is the complete path of a request:

**External Traffic** ➡️ `http://app.mydomain.com`

**Ingress** ➡️ Sees the rule and routes the traffic to the `my-service` Service.

**Service** ➡️ Receives the traffic on its `port: 80`.

**Pod** ➡️ The Service forwards the traffic to the `targetPort: 3000` of one of the selected Pods.

**Container** ➡️ Your application, listening on port 3000, receives the request.

Here are the four main Kubernetes Service types.

### ClusterIP

This is the **default** Service type. It exposes the Service on an internal-only IP address that is only reachable from
within the cluster.

- **Use Case:** For internal communication between different microservices within your cluster (e.g., a web frontend
  talking to a backend API).
- **Analogy:** An internal extension number in an office phone system.

### NodePort

This exposes the Service on a static port on the IP address of **every node** in the cluster.

- **Use Case:** For exposing an application for development or testing purposes when you don't have a cloud load
  balancer. You access it using `<NodeIP>:<NodePort>`.
- **Analogy:** A fire escape on an apartment building. It's a direct, less elegant way to get in from the outside.

### LoadBalancer

This is the standard way to expose a service to the internet. It's an extension of `NodePort`.

- **Use Case:** On a cloud provider (like AWS, GCP, Azure), this automatically provisions a cloud load balancer, which
  then directs external traffic to your service's `NodePort`.
- **Analogy:** The official front entrance and lobby of an office building. It manages and directs all incoming public
  traffic.

### ExternalName

This is a special case. Instead of creating an internal IP, it acts as a DNS alias.

- **Use Case:** To give a service inside your cluster a stable name that points to an **external service** (like a
  managed database or a third-party API).
- **Analogy:** A mailing address forwarding service. Requests sent to `my-db.my-cluster` are automatically forwarded to
  an external address like `rds.amazonaws.com`.

### Headless Service

A variation of a `ClusterIP` Service, but it does not have a stable IP address for load balancing. Instead, it provides
a way for you to discover the individual IP addresses of all the pods it selects.

When you perform a DNS lookup on a normal service, you get back one IP—the service's `ClusterIP`. When you do a DNS
lookup on a **Headless Service**, you get back a **list of the IP addresses of all the individual pods** that match the
service's selector.

This is primarily used for stateful applications where direct communication between pods is necessary.

- **Peer Discovery:** In a distributed database or clustered application (like Zookeeper, Cassandra, or etcd), each pod
  (peer) needs to know the IP addresses of the other pods to form a quorum, replicate data, and communicate directly.
- **StatefulSets:** Headless Services are almost always used with `StatefulSets`. This combination provides each pod
  with a stable, unique network identity (e.g., `my-db-0.my-headless-service`, `my-db-1.my-headless-service`) that other
  pods can use to find it.

Analogy:

- **Normal Service (`ClusterIP`):** Like calling a company's main switchboard. You get connected to _any_ available
  operator.
- **Headless Service:** Like looking up the company directory and getting the **direct-dial number** for _every single
  operator_, allowing you to call a specific one directly.

Example YAML:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-headless-service
spec:
  clusterIP: None # This is the key that makes it headless
  selector:
    app: my-stateful-app
  ports:
    - port: 80
```

## DNS in Kubernetes

Kubernetes uses CoreDNS on port 53 for DNS resolution. CoreDNS is a flexible, extensible DNS server.

Kubernetes resources for coreDNS are:

a service account named coredns,

cluster-roles named coredns and kube-dns

clusterrolebindings named coredns and kube-dns,

a deployment named coredns,

a configmap named coredns and a

service named kube-dns.

While analyzing the coreDNS deployment you can see that the the Corefile plugin consists of important configuration
which is defined as a configmap.

    kubernetes cluster.local in-addr.arpa ip6.arpa {
       pods insecure
       fallthrough in-addr.arpa ip6.arpa
       ttl 30
    }

This is the backend to k8s for cluster.local and reverse domains.

proxy . /etc/resolv.conf

Forward out of cluster domains directly to right authoritative DNS server.

## Gateway API

The kubernetes gateway API focuses on layer 4 and 7 routing. It defines custom resources, but a controller is needed to
actually implement them.

We'll use the NGINX Gateway Controller, which supports all standard Gateway API resources.

To install the NGINX Gateway Controller, run the following commands:

```bash
kubectl kustomize "https://github.com/nginx/nginx-gateway-fabric/config/crd/gateway-api/standard?ref=v1.6.2" | kubectl apply -f -
```

Or with Helm:

```bash
helm install ngf oci://ghcr.io/nginx/charts/nginx-gateway-fabric --create-namespace -n nginx-gateway
```

### GatewayClass

A GatewayClass defines a set of Gateways that are implemented by a specific controller. Think of it as a blueprint that
tells Kubernetes which controller will manage the Gateways.

Purpose: Decouples Gateway configuration from the actual implementation: This allows you to define Gateways without
worrying about the underlying controller.

It supports multiple gateway implementations in a single cluster (e.g you can have both NGINX and Istio Gateways in the
same cluster).

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: nginx-gateway-class
  namespace: default
spec:
  controllerName: example.com/gateway-controller
```

### Configuring HTTP Gateway and Listener

A Gateway is a Kubernetes resource that defines how traffic enters your cluster. It specifies the protocols, ports, and
routing rules for incoming traffic.

Here's an example of a Gateway that listens for HTTP traffic on port 80 and forward it to the appropriate backend
services:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: nginx-gateway
  namespace: default
spec:
  gatewayClassName: nginx
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
```

`gatewayClassName`: Refers to the GatewayClass (e.g., nginx) that will manage this Gateway.

`listeners`: Defines how the Gateway listens for traffic.

`name`: A unique name for this listener.

`protocol`: Specifies that this listener will handle HTTP traffic.

`port`: The port number on which the Gateway will listen for HTTP traffic.

`allowedRoutes`: Specifies which namespaces can define routes for this Gateway. Here, from: All allows routes from all
namespaces.

### HTTP Routing

An `HTTPRoute` defines how HTTP traffic is forwarded to Kubernetes services. It works in conjunction with a Gateway to
route requests based on specific rules, such as matching paths or headers.

Here's an example of an HTTPRoute:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: basic-route
  namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
  rules:
    - matches:
        - path:
          value: /app
          type: PathPrefix
      backendRefs:
        - name: my-app
          port: 80
```

`parentRefs`: Links this route to a specific Gateway (e.g., nginx-gateway).

`rules`: Defines how traffic is routed.

`matches`: Specifies the conditions for matching traffic.

`path`: Matches requests with a specific path prefix (e.g., /app).

`backendRefs`: Specifies the backend service and port to which the traffic should be forwarded.

This configuration routes all requests with the path prefix /app to my-app service on port 80.

[HTTP Routing Guide](https://gateway-api.sigs.k8s.io/guides/http-routing/)

### HTTP Redirects and Rewrites

Redirects and rewrites are powerful tools for modifying incoming requests before they reach the backend service.

**Example: HTTP to HTTPS**. Redirect Redirects are used to force traffic to a different scheme (e.g., HTTP to HTTPS):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: https-redirect
  namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
    rules:
      - filters:
          - type: RequestRedirect
            requestRedirect:
              scheme: https
```

`filters`: Defines additional processing for requests.

`type`: RequestRedirect: Specifies that this filter will redirect requests.

`requestRedirect.scheme`: Redirects all HTTP requests to HTTPS.

This configuration ensures that all incoming HTTP traffic is redirected to HTTPS, improving security.

[HTTP Redirects Guide](https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/)

**Example: Path Rewrite**. Rewrites modify the request path before forwarding it to the backend:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: rewrite-path
  namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /old
      filters:
        - type: URLRewrite
          urlRewrite:
            path:
              replacePrefixMatch: /new
      backendRefs:
        - name: my-app
          port: 80
```

`matches.path`: Matches requests with the path prefix /old.

`filters.type`: URLRewrite: Specifies that this filter will rewrite the URL.

`replacePrefixMatch: /new`: Replaces the /old prefix with /new.

`backendRefs`: Forwards the modified request to my-app service on port 80.

This configuration rewrites requests from /old to /new before sending them to the backend.

[HTTP Rewrite Guide](https://gateway-api.sigs.k8s.io/guides/http-redirect-rewrite/)

### HTTP Header Modification

You can modify HTTP headers in requests or responses to add, set, or remove specific headers.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: header-mod
  namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
  rules:
    - filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            add:
              x-env: staging
      backendRefs:
        - name: my-app
          port: 80
```

`filters.type`: RequestHeaderModifier: Specifies that this filter will modify request headers.

`add.x-env`: Adds a custom header (x-env) with the value staging.

`backendRefs`: Forwards the modified request to the my-app service on port 80.

This configuration is useful for adding metadata to requests, such as environment-specific headers.

[HTTP Header Guide](https://gateway-api.sigs.k8s.io/guides/http-header-modifier/)

### HTTP Traffic Splitting

Traffic splitting allows you to distribute traffic between multiple backend services. This is often used for canary
deployments or A/B testing.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: traffic-split
  namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
  rules:
    - backendRefs:
        - name: v1-service
          port: 80
          weight: 80
        - name: v2-service
          port: 80
          weight: 20
```

`backendRefs`: Specifies the backend services and their weights.

`weight: 80`: Sends 80% of traffic to v1-service.

`weight: 20`: Sends 20% of traffic to v2-service.

This configuration splits traffic between two services, with most traffic going to v1-service.

[HTTP Traffic Splitting Guide](https://gateway-api.sigs.k8s.io/guides/traffic-splitting/)

### HTTP Request Mirroring

Request mirroring allows you to send a copy of incoming requests to a secondary service for testing or analysis, without
affecting the primary service.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: request-mirror
  namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
      rules:
    - filters:
        - type: RequestMirror
          requestMirror:
            backendRef:
              name: mirror-service
              port: 80
      backendRefs:
        - name: my-app
          port: 80
```

`filters.type`: RequestMirror: Specifies that this filter will mirror requests.

`requestMirror.backendRef`: Points to the secondary service mirror-service that will receive the mirrored requests.

`backendRefs`: Forwards the original request to the primary service my-app.

This configuration is useful for testing new services or analyzing traffic patterns without impacting production.

[HTTP Traffic Request Guide](https://gateway-api.sigs.k8s.io/guides/http-request-mirroring/)

### TLS Configuration

TLS (Transport Layer Security) is used to encrypt traffic between clients and servers, ensuring secure communication. In
Kubernetes, you can terminate TLS traffic at the Gateway level by using a certificate stored in a Kubernetes Secret.
This means the Gateway will handle decrypting the traffic before forwarding it to backend services.

**Example: TLS Termination**. The following example demonstrates how to configure a Gateway to terminate TLS traffic:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: nginx-gateway-tls
  namespace: default
spec:
  gatewayClassName: nginx
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: tls-secret
      allowedRoutes:
        namespaces:
          from: All
```

`protocol`: Specifies that this listener will handle HTTPS traffic.

`tls.mode`: Indicates that the Gateway will terminate the TLS connection (decrypt the traffic).

`certificateRefs`: Points to a Kubernetes Secret (e.g., tls-secret) that contains the TLS certificate and private key.

`allowedRoutes`: Configures which namespaces can define routes for this Gateway. Here, `from: All` allows routes from
all namespaces.

This setup is commonly used for secure communication between clients and the Gateway, while backend services receive
unencrypted traffic.

[TLS Configuration Guide](https://gateway-api.sigs.k8s.io/guides/tls/)

### TCP, UDP, and Other Protocols

The Gateway API supports more than just HTTP traffic. You can configure Gateways to handle protocols like TCP, UDP, and
even gRPC. This flexibility makes it suitable for a wide range of applications, such as databases, DNS servers, and
microservices.

**TCP Example**

TCP is a connection-oriented protocol often used for applications like databases. The following example shows how to
configure a Gateway for TCP traffic:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: tcp-gateway
  namespace: default
spec:
  gatewayClassName: nginx
  listeners:
    - name: tcp
      protocol: TCP
      port: 3306
      allowedRoutes:
      namespaces:
      from: All
```

`protocol`: Specifies that this listener will handle TCP traffic.

`port`: The port number for the listener, commonly used for MySQL databases.

`allowedRoutes`: Allows routes from all namespaces to use this Gateway.

This configuration is ideal for exposing database services to external clients.

**UDP Example**

UDP is a connectionless protocol often used for DNS or streaming applications. Here's an example of a Gateway configured
for UDP traffic:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: udp-gateway
  namespace: default
spec:
  gatewayClassName: nginx
  listeners:
    - name: udp
      protocol: UDP
      port: 53
      allowedRoutes:
      namespaces:
      from: All
```

`protocol`: Specifies that this listener will handle UDP traffic.

`port`: The port number for the listener, commonly used for DNS services.

`allowedRoutes`: Allows routes from all namespaces to use this Gateway.

This setup is useful for exposing DNS services or other UDP-based applications.

**gRPC Example**

gRPC is a high-performance RPC (Remote Procedure Call) framework often used in microservices. The Gateway API supports
gRPC by using HTTPRoute resources. Here's an example:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
name: grpc-route
namespace: default
spec:
  parentRefs:
    - name: nginx-gateway
  rules:
    - matches:
        - method:
            service: my.grpc.Service
            method: GetData
      backendRefs:
        - name: grpc-service
          port: 50051
```

`method.service`: Specifies the gRPC service name (e.g., my.grpc.Service).

`method.method`: Specifies the gRPC method to match (e.g., GetData).

`backendRefs`: Points to the backend service (grpc-service) and its port 50051.

This configuration routes gRPC requests to the appropriate backend service, enabling seamless communication between
microservices.

## Kube-Proxy

kube-proxy is a network proxy that runs and maintains network rules on each node in the cluster. These rules allow
network communication to the Pods from network sessions inside or outside of the cluster.

kube-proxy is responsible for watching services and endpoint associated with each service. When the client is going to
connect to the service using the virtual IP, kube-proxy is responsible for sending traffic to actual pods.

In a cluster configured with kubeadm, you can find kube-proxy as a daemonset.

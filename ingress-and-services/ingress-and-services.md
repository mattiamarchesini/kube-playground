# Ingress and services

## Complete Flow Summary

Here is the complete path of a request:

**External Traffic** ➡️ `http://app.mydomain.com`

**Ingress** ➡️ Sees the rule and routes the traffic to the `my-service` Service.

**Service** ➡️ Receives the traffic on its `port: 80`.

**Pod** ➡️ The Service forwards the traffic to the `targetPort: 3000` of one of the selected Pods.

**Container** ➡️ Your application, listening on port 3000, receives the request.

## Service types

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

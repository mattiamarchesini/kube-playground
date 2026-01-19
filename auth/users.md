# Users

To create a new user after the admin user, you can use the certificate API. After a CertificateSigninRequest object is
created, the request can be approved and reviewed easily using `kubectrl`.

First, the user Jane creates a key and a certificate singing request:

```bash
openssl genrsa -out jane.key 2048

openssl req -new -key jane.key -subj "/CN=Jane" -out jane.csr
```

Jane then sends the request to the admin, who will use it to create a signing request object as seen in
[`certificate-signing-request.yaml`](./certificate-signing-request.yaml) and approve it using

```bash
kubectl certificate approve jane-csr
```

Now Jane can check what she can do using the `auth can-i` subcommand, for example:

```bash
kubectl auth can-i create deployments

kubectl auth can-i create deployments --namespace dev

kubectl auth can-i create deployments --as <different-user>
```

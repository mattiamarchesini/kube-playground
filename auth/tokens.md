# Token

Tokens are managed and rotated automatically by kubernetes when a new service account is created.

To manually create a new token (encoded base 64), use:

```bash
kubectl create token a-token --duration 2h
```

You can use the token to call the kubernetes api:

```bash
curl --insecure https://server.com/api --header "Authorization: Bearer <TOKEN>"
```

# Security

## Security primitives

- root access disabled
- password based auth disabled (only ssh access keys)

Controlling access to the api-server

- Certificates
- Tokens
- Service accounts: Used by machines
  - every namespace has a default service account
  - on attach, a new token is mounted to the pod
- External auth providers
  - LDAP

What a user can do is defined by authentication modes:

- Node Auth: the one used by nodes to communicate
- ABAC Auth: Attribute-Based
- RBAC Auth: Role-Based
- Webhook mode: for webhooks
- AlwaysAllow
- AlwaysDeny

All communication between cluster components is secured using TLS encryption. By default, all pods can communicate with
each other. You can restrict access between them using network policies.

## Authentication

[SSL/TLS Explained in 7 Minutes](https://www.youtube.com/watch?v=67Kfsmy_frM)

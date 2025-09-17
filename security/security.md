# Security

## Security primitives

- root access disabled
- password based auth disabled (only ssh access keys)

Controlling access to the api-server

- static token files
- certificates
- external auth providers
  - LDAP
- service accounts

What can a user do is defined by:

- RBAC Auth
- ABAC Auth
- Node Auth
- Webhook mode

All communication between cluster components is secured using TLS encryption. By default, all pods can communicate with
each other. You can restrict access between them using network policies.

## Authentication

[SSL/TLS Explained in 7 Minutes](https://www.youtube.com/watch?v=67Kfsmy_frM)

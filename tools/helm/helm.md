# Helm

Helm acts as a package manager for Kubernetes. It bundles all the YAML files needed for an application (Deployments,
Services, etc.) into a single, reusable package called a Chart. With one command, Helm lets you install, upgrade, and
manage entire applications, instead of applying dozens of YAML files manually.

```bash
helm install wordpress
helm install mysite01 wordpress
helm upgrade wordpress
helm rollback wordpress
```

You can even configure your applications with a yaml file:

```yaml
# values.yaml
wordpressUsername: user
wordpressEmail: user@example.com
wordpressBlogName: User's Blog
```

## Install

Run

```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

or use your system package manager.

## Charts

Charts files contain all the information needed to create the objects. They're essentially templates to fill with the
values found in `values.yaml` or passed by the flags `--set` or `-f`.

Here's how a deployment would look like:

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: "{{ .Values.appName }}"
spec:
  replicas: { { .Values.replicaCount } }
  selector:
    matchLabels:
      app: "{{ .Values.appName }}"
# ...
```

Helm charts use the Go templating language ([Go templates](https://pkg.go.dev/text/template)), to which Helm adds a set
of special objects and functions for managing Kubernetes.

The file [Chart.yaml](./Chart.yaml) at the root of the project contains info about the package itself.

This is the structure tree of a typical Helm chart project:

```
├── charts
│   ├── dependency1.tgz
│   ├── dependency2
|   |   ├── Chart.yaml
│   |   └── ...
│   └── ...
├── Chart.yaml
├── LICENSE
├── README.md
├── templates
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── _helpers.tpl
│   ├── _another-component.tpl
│   ├── ...
│   └── NOTES.txt
└── values.yaml
```

Templates have the extension `.tpl` as a convention to let your tools know it's supposed to be a template and not a
complete yaml definitions. Their file names also need to start with `_` to let Helm know they won't be part of the
final output.

To output the YAML from a chart template into a file use:

```bash
helm template my-release bitnami/nginx -f my-values.yaml > my-nginx.yaml
```

When a chart is installed to the cluster, a release is created. A release can have multiple revisions, which are
snapshots of the application.

Helm saves all the metadata needed to operate as secrets on the cluster in order to enable collaboration.

Charts from different sources and vendors are hosted at https://artifacthub.io/. You can also add a chart repo using the
`helm repo add <repo name> <repo url>` command. To search for charts use `helm search [hub, repo] <name>`.

To list available charts use `helm list`. Use `helm history <chart>` to list revisions.

When installing an app, you can override values with custom parameters like this:

```bash
helm install --set appName="myapp" --set replicaCount=2

# Or
helm install --values custom-values.yaml

# Or
helm pull --untar bitnami/wordpress
# change values
sed -i values.yaml 's/value01: 1/value01: 2/g'
# install
helm install my-wordpress-relese ./wordpress
```

You can install a specific version using

```bash
helm install nginx-release bitnami/nginx --version 7.1.0
```

To upgrade to the new version

```bash
helm upgrade nginx-release bitnami/nginx
```

To upgrade to a specific version of the image and helm chart:

```bash
helm upgrade nginx-release bitnami/nginx --version 18.3.6 --set image.tag=1.27.1
```

To rollback to a precedent release (rollback actually creates a new release):

```bash
helm rollback nginx-release <release number>
```

**Chart hooks** can be used to perform administrative operations during the lifecycle of the application (e.g. take a
backup of the DB before performing an update). Here are all the hooks:

- pre/post-install
- pre/post-upgrade
- pre/post-delete
- pre/post-rollback
- test: when using `helm test`

Here's an example of how to implement an hook:

```yaml
# templates/db-migration-job.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: "{{ .Release.Name }}-db-migration"
  labels:
    app: { { .Chart.Name } }
  # HOOK Section
  annotations:
    # Specify the points in the lifecycle at which to run this Job
    "helm.sh/hook": "pre-install, pre-upgrade"

    # Specify the order (lowest numbers are executed first)
    "helm.sh/hook-weight": "-5"

    # Specifies what to do with the Job after it has been successful
    "helm.sh/hook-delete-policy": "hook-succeeded"
spec:
  template:
    spec:
      containers:
        - name: db-migration
          image: "my-app/migrator:{{ .Chart.AppVersion }}"
          # The command that performs the migration
          command: ["/run_migration.sh"]
      restartPolicy: Never
  backoffLimit: 1
```

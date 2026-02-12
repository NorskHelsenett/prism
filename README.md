# PRISM - Pentest Report Information Security Management

PRISM is an internal vulnerability management and reporting tool developed by **CAT (Cyber Action Team)**, the internal penetration testing team at **Norsk helsenett SF (NHN)**. It provides a centralized platform for delegating, tracking, and reporting security vulnerabilities discovered during assessments.

## Screenshots

<table>
  <tr>
    <td><img src=".docs/login.png" alt="Login Screen" width="100%"/></td>
    <td><img src=".docs/frontpage.png" alt="Front Page" width="100%"/></td>
  </tr>
  <tr>
    <td><img src=".docs/projects.png" alt="Projects Dashboard" width="100%"/></td>
    <td><img src=".docs/vulnerabilities.png" alt="Vulnerabilities List" width="100%"/></td>
  </tr>
  <tr>
    <td><img src=".docs/project.png" alt="Project Details" width="100%"/></td>
    <td><img src=".docs/vulnerability.png" alt="Vulnerability Details" width="100%"/></td>
  </tr>
</table>

## Architecture

PRISM consists of two main components:

- **API** - Backend service written in Go, providing a GraphQL and REST API
- **Web** - Frontend application built with SvelteKit

## Development Setup

### Using DevContainer (Recommended)

The easiest way to get started is using the provided DevContainer configuration with VS Code or any compatible IDE.

1. Open the project in VS Code
2. When prompted, click "Reopen in Container" (or run the command `Dev Containers: Reopen in Container`)
3. The container will automatically:
   - Install Go, Node.js, and required dependencies
   - Start [mocc](https://github.com/jonasbg/mocc) as a local OIDC provider on port 9999
   - Install npm dependencies and start the web dev server

4. Start the API server in a terminal:
   ```bash
   cd api
   CONFIG_PATH="config.yaml" go run .
   ```

5. Access the application at `http://localhost:5173`

#### Test Users (via mocc)

The mocc OIDC provider comes with pre-configured test users:

| Email | Role |
|-------|------|
| `alice.admin@test.local` | Admin (configured in config.yaml) |
| `bob.user@test.local` | Regular user |
| `charlie.viewer@test.local` | Regular user |

### Manual Setup

If not using DevContainer:

1. Install Go 1.25+ and Node.js 20+
2. Start a mocc instance or configure another OIDC provider
3. Create `api/config.yaml` (see Configuration section)
4. Install and run:
   ```bash
   # Terminal 1 - API
   cd api
   CONFIG_PATH="config.yaml" go run .

   # Terminal 2 - Web
   cd web
   npm install
   npm run dev
   ```

## Configuration

PRISM is configured via a YAML file. The path is specified by the `CONFIG_PATH` environment variable.

### Example config.yaml

```yaml
oidc:
  mocc:                                          # Provider key (used in URL path)
    name: "Mocc IdP"                             # Display name on login page
    clientID: "prism-local-client"
    clientSecret: "prism-local-secret"
    redirectUri: "http://localhost:5173/api/callback"
    providerUri: "http://localhost:9999"

cors:
  origin: "http://localhost:5173"                # Frontend URL

admins:
  - alice.admin@test.local                       # Users with admin privileges

database:
  path: "./.tmp"                                 # SQLite database location

events:
  interval: 60                                   # Event processing interval (seconds)

slack:
  token: ""                                      # Slack bot token (optional)
  webhookUrl: ""                                 # Slack webhook URL (optional)

secrets:
  HMAC_SECRET_KEY: "change-this-in-production"  # Used for signing tokens
```

### Multiple OIDC Providers

You can configure multiple OIDC providers:

```yaml
oidc:
  azure:
    name: "Microsoft AD"
    clientID: "your-azure-client-id"
    clientSecret: "your-azure-client-secret"
    redirectUri: "https://prism.example.com/api/callback"
    providerUri: "https://login.microsoftonline.com/your-tenant-id/v2.0"
  gitlab:
    name: "GitLab"
    clientID: "your-gitlab-client-id"
    clientSecret: "your-gitlab-client-secret"
    redirectUri: "https://prism.example.com/api/callback"
    providerUri: "https://gitlab.example.com"
```

## Kubernetes Deployment

PRISM is distributed as a Helm chart via `ghcr.io/norskhelsenett/prism/helm`.

### Prerequisites

- Kubernetes cluster
- Helm 3+

### 1. Create the Configuration Secret

The API configuration must be deployed as a Kubernetes Secret before installing the Helm chart:

```bash
kubectl create secret generic api-config-secret \
  --from-file=config.yaml=/path/to/your/config.yaml \
  -n your-namespace
```

Example production `config.yaml`:

```yaml
oidc:
  azure:
    name: "Microsoft Entra ID"
    clientID: "your-client-id"
    clientSecret: "your-client-secret"
    redirectUri: "https://prism.example.com/api/callback"
    providerUri: "https://login.microsoftonline.com/your-tenant-id/v2.0"

cors:
  origin: "https://prism.example.com"

admins:
  - admin@example.com

database:
  path: "/data"

events:
  interval: 60

slack:
  token: "xoxb-your-slack-token"
  webhookUrl: "https://hooks.slack.com/services/xxx/yyy/zzz"

secrets:
  HMAC_SECRET_KEY: "generate-a-secure-random-string"
```

### 2. Create values-override.yaml

```yaml
image:
  repository: ghcr.io/norskhelsenett/prism
  tag: "latest"

prism:
  ingress:
    enabled: true
    host: prism.example.com
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    tls: true

  storageClassName: standard

  # Resource limits (adjust based on your needs)
  resources:
    requests:
      memory: "256Mi"
      cpu: "100m"
    limits:
      memory: "1Gi"
      cpu: "1"

  # The config secret must exist before deployment
  mounts:
    - name: config
      mountPath: /config
      fromSecret: api-config-secret
      readOnly: true
    - name: data
      mountPath: /data
      storage: 10Gi
      statefulSet: true
    - name: tmp
      inMemory: true
      mountPath: /tmp
      storage: 50Mi
```

### 3. Install the Chart

```bash
helm upgrade --install prism oci://ghcr.io/norskhelsenett/prism/helm \
  -f values-override.yaml \
  -n your-namespace \
  --create-namespace
```

### 4. Verify Deployment

```bash
kubectl get pods -n your-namespace
kubectl logs -f prism-0 -n your-namespace
```

## About mocc

[mocc](https://github.com/jonasbg/mocc) (Minimal OpenID Connect Core) is a tiny, opinionated mock OIDC provider written in Go. It supports the authorization code flow, provides a JWKS endpoint, and issues short-lived ID tokens. The DevContainer automatically starts mocc on port 9999.

## Project Structure

```
prism/
├── api/                    # Go backend
│   ├── auth/               # OIDC authentication
│   ├── config/             # Configuration loading
│   ├── database/           # SQLite database layer
│   ├── models/             # Data models
│   ├── routes/             # HTTP handlers
│   └── main.go             # Entry point
├── web/                    # SvelteKit frontend
│   ├── src/
│   │   ├── lib/            # Shared components and utilities
│   │   └── routes/         # Page routes
│   └── static/             # Static assets
├── .cluster/               # Helm charts
│   └── prism/
├── .devcontainer/          # DevContainer configuration
└── .docs/                  # Documentation and screenshots
```

## Security

Report security vulnerabilities to cat[at]nhn.no. See [.well-known/security.txt](.well-known/security.txt) for details.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Maintainers

**CAT - Cyber Action Team**
Norsk helsenett SF (NHN)

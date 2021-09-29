## apiserver
Kuberlogic provides a first class REST API to manage services, backup configurations and more.

## Configuration
In order to integrate smoothly into Kuberlogic, `apiserver` needs to be configured correctly. It is configured via environment variables:

| name | type | default | description |
| --- | --- | --- | --- |
| KUBERLOGIC_BIND_HOST | string | 0.0.0.0 | A host to listen on. Required |
| KUBERLOGIC_HTTP_BIND_PORT | int | 8081 | A port to listen on. Required |
| KUBERLOGIC_AUTH_PROVIDER | string | “” | Authentication provider for the REST interface. Supported: “keycloak” | “none” | Optional.
| KUBERLOGIC_AUTH_KEYCLOAK_CLIENT_ID | string | “” | Keycloak client ID for “keycloak” authentication provider. Optional. |
| KUBERLOGIC_AUTH_KEYCLOAK_CLIENT_SECRET | string | “” | Keycloak client secret. Optional. |
| KUBERLOGIC_AUTH_KEYCLOAK_REALM_NAME | string | “” | Keycloak realm name. Optional. |
| KUBERLOGIC_AUTH_KEYCLOAK_URL | string | “” | Keycloak URL. Optional. |
| KUBERLOGIC_KUBECONFIG_PATH | string | “/root/.kube/config
” |  Absolute path to kubeconfig when started outside of Kubernetes cluster. Optional. |
|  KUBERLOGIC_DEBUG_LOGS | bool | false | Enable debug logging. Optional. |
| SENTRY_DSN | string | “” | Sentry URL to report panics. Optional. |
| POSTHOG_AP_KEY | string | “” | Posthog API key for statistics. Optional. |
| KUBERLOGIC_CORS_ALLOWED_ORIGINS | string | “https://*;http://*” | `;` separated list of CORS allowed origins. Optional. |

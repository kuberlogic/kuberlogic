## operator
Kuberlogic’s heart is `operator`. It is responsible for keeping services in a healthy state.

## Configuration
`operator` needs a set of configuration parameters passed as environment variables:

| name | type | default | description |
| --- | --- | --- | --- |
| IMG_REPO | string | “” | Container image repository where Kuberlogic container images are located. Required. |
| IMG_PULL_SECRET | string | “” | ImagePullSecret name for the registry of Kuberlogic container images. Optional. |
| POD_NAMESPACE | string | “” | Namespace in which Kuberlogic `operator` is running. Required. |
| SENTRY_DSN | string | “” | Sentry URL to report panics. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_ENABLED | bool | false | Enable email notification channel for Kuberlogic alerts. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_HOST | string | “” | SMTP host for email notification channel. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_PORT | int | 0 | SMTP port for . Optional. |
| NOTIFICATION_CHANNELS_EMAIL_TLS_INSECURE | bool | false | Do not verify TLS when connected to SMTP server. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_TLS_ENABLED | bool | false | Use TLS when connecting to SMTP server. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_USERNAME | string | “” | SMTP server connection username. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_PASSWORD | string | “” | SMTP server connection password. Optional. |
| NOTIFICATION_CHANNELS_EMAIL_FROM | string | "operator@example.com" | `From:` address for email notifications. Optional. |
| GRAFANA_ENABLED | bool | false | Enable Grafana integration for Kuberlogic operator. Optional. |
| GRAFANA_ENDPOINT | string | “” | Grafana URL. Optional. |
| GRAFANA_LOGIN | string | “” | Grafana admin username. Optional. |
| GRAFANA_PASSWORD | string | “” | Grafana admin password. Optional. |
| GRAFANA_DEFAULT_DATASOURCE_ENDPOINT | “” | Prometheus URL to configure as a Grafana main datasource. Optional. |
| PLATFORM | string | "GENERIC" | A platform where Kuberlogic runs. Optional. |
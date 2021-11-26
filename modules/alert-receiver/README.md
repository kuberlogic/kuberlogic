## alert-receiver
`alert-receiver` is a fairly simple application that is designed to translate Prometheus' alert-manager webhook alerts notifications into KuberLogic's `kuberlogicalert` objects.
`alert-receiver` is deployed alongside with an `alert-manager` and waits for webhook requests with an alert payload.

## Alerts processing
`alert-receiver` expects to receive a specific set of alerts:
* `service` scoped: alerts based on KuberLogic metrics
* `instance` scoped: alerts based on Kubernetes metrics

Scope is determined by checking the `annotations.kuberlogicAlertScope` field of an alert.

Additionally, an alert must have these labels set:
* resourcename: name of the resource that caused an alert
* namespace: namespace where the instance or service are deployed
* text: text message that describes an alert
* severity: alert severity (medium | high).

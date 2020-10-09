package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	cmStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_ready",
			Help: "CloudManaged application ready status",
		},
		labelList,
	)
	cmReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_replicas",
			Help: "CloudManaged application replicas",
		},
		labelList,
	)
	cmMemLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_memory_limit_bytes",
			Help: "CloudManaged application memory limit",
		},
		labelList,
	)
	cmCPULimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_cpu_limit_milliseconds",
			Help: "CloudManaged application cpu limit",
		},
		labelList,
	)
)

// Init Registers Prometheus metrics in global metrics registry provided by controller-runtime
func Init() error {
	metrics.Registry.MustRegister(
		cmStatus,
		cmReplicas,
		cmMemLimit,
		cmCPULimit,
	)

	return nil
}

func exposeReadinessMetric(ready bool, labels map[string]string) {
	notReadyVal := 0
	if ready {
		notReadyVal = 1
	}

	cmStatus.With(labels).Set(float64(notReadyVal))
}

func exposeReplicasMetric(replicas int32, labels map[string]string) {
	cmReplicas.With(labels).Set(float64(replicas))
}

func exposeMemLimitMetric(memLimit int64, labels map[string]string) {
	cmMemLimit.With(labels).Set(float64(memLimit))
}

func exposeCPULimitsMetric(cpuLimit int64, labels map[string]string) {
	cmCPULimit.With(labels).Set(float64(cpuLimit))
}

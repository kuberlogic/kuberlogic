package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var metricLabels = []string{
	"Name",
	"Namespace",
	"type",
}

var (
	cmStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_ready",
			Help: "CloudManaged application ready status",
		},
		metricLabels,
	)
	cmReplicas = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_replicas",
			Help: "CloudManaged application replicas",
		},
		metricLabels,
	)
	cmMemLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_memory_limit_bytes",
			Help: "CloudManaged application memory limit",
		},
		metricLabels,
	)
	cmCPULimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cloudmanaged_cpu_limit_milliseconds",
			Help: "CloudManaged application cpu limit",
		},
		metricLabels,
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

type MetricsMetadata struct {
	Name        string
	Namespace   string
	ClusterType string
}

func PopulateStatusMetric(meta *MetricsMetadata, ready bool) {
	notReadyVal := 0
	if ready {
		notReadyVal = 1
	}

	cmStatus.With(prometheus.Labels{
		"Name":      meta.Name,
		"Namespace": meta.Namespace,
		"type":      meta.ClusterType,
	}).Set(float64(notReadyVal))
}

func PopulateReplicasMetric(meta *MetricsMetadata, replicas int32) {
	cmReplicas.With(prometheus.Labels{
		"Name":      meta.Name,
		"Namespace": meta.Namespace,
		"type":      meta.ClusterType,
	}).Set(float64(replicas))
}

func PopulateMemLimitMetric(meta *MetricsMetadata, memLimit int64) {
	cmMemLimit.With(prometheus.Labels{
		"Name":      meta.Name,
		"Namespace": meta.Namespace,
		"type":      meta.ClusterType,
	}).Set(float64(memLimit))
}

func PopulateCPULimitsMetric(meta *MetricsMetadata, cpuLimit int64) {
	cmCPULimit.With(prometheus.Labels{
		"Name":      meta.Name,
		"Namespace": meta.Namespace,
		"type":      meta.ClusterType,
	}).Set(float64(cpuLimit))
}

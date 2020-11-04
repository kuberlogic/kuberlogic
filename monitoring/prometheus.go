package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
)

var labels = []string{
	"name",
	"namespace",
	"cluster_type",
}

var (
	cmReady = prometheus.NewDesc(
		"cloudmanaged_ready",
		"CloudManaged application ready status", labels, nil)
	cmReplicas = prometheus.NewDesc(
		"cloudmanaged_replicas",
		"CloudManaged application replicas", labels, nil)
	cmMemLimit = prometheus.NewDesc(
		"cloudmanaged_memory_limit_bytes",
		"CloudManaged application memory limit", labels, nil)
	cmCPULimit = prometheus.NewDesc(
		"cloudmanaged_cpu_limit_milliseconds",
		"CloudManaged application cpu limit", labels, nil)
)

var CloudManageds = make(map[string]*cloudlinuxv1.CloudManaged)

// Implements prometheus.Collector
type CloudManagedCollector struct {
}

func (c CloudManagedCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cmReady
	ch <- cmReplicas
	ch <- cmMemLimit
	ch <- cmCPULimit
}

func (c CloudManagedCollector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range CloudManageds {
		ch <- prometheus.MustNewConstMetric(
			cmReady, prometheus.GaugeValue, calcReadinessMetric(c),
			c.Name, c.Namespace, c.Spec.Type)
		ch <- prometheus.MustNewConstMetric(
			cmReplicas, prometheus.GaugeValue, float64(c.Spec.Replicas),
			c.Name, c.Namespace, c.Spec.Type)
		ch <- prometheus.MustNewConstMetric(
			cmMemLimit, prometheus.GaugeValue, float64(c.Spec.Resources.Limits.Memory().Value()),
			c.Name, c.Namespace, c.Spec.Type)
		ch <- prometheus.MustNewConstMetric(
			cmCPULimit, prometheus.GaugeValue, float64(c.Spec.Resources.Limits.Cpu().MilliValue()),
			c.Name, c.Namespace, c.Spec.Type)
	}
}

func calcReadinessMetric(cm *cloudlinuxv1.CloudManaged) float64 {
	if cm.Status.Status == cloudlinuxv1.ClusterOkStatus {
		return float64(1)
	} else {
		return float64(0)
	}
}

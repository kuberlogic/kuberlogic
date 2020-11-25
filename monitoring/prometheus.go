package monitoring

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
)

var labels = []string{
	"name",
	"namespace",
	"cluster_type",
}

var backupLabels = []string{
	"name",
	"namespace",
	"cluster",
}

var (
	cmReady = prometheus.NewDesc(
		"cloudmanaged_ready",
		"CloudManaged application ready status",
		labels,
		nil)
	cmReplicas = prometheus.NewDesc(
		"cloudmanaged_replicas",
		"CloudManaged application replicas",
		labels,
		nil)
	cmMemLimit = prometheus.NewDesc(
		"cloudmanaged_memory_limit_bytes",
		"CloudManaged application memory limit",
		labels,
		nil)
	cmCPULimit = prometheus.NewDesc(
		"cloudmanaged_cpu_limit_milliseconds",
		"CloudManaged application cpu limit",
		labels,
		nil)

	// Backup
	cmBackupSuccess = prometheus.NewDesc(
		"cloudmanagedbackup_success",
		"CloudManaged backup success",
		backupLabels,
		nil)
	cmBackupStatus = prometheus.NewDesc(
		"cloudmanagedbackup_status",
		"CloudManaged backup status",
		backupLabels,
		nil)
	// Restore
	cmRestoreSuccess = prometheus.NewDesc(
		"cloudmanagedrestore_success",
		"CloudManaged restore success",
		backupLabels,
		nil)
	cmRestoreStatus = prometheus.NewDesc(
		"cloudmanagedrestore_status",
		"CloudManaged restore status",
		backupLabels,
		nil)
)

var CloudManageds = make(map[string]*cloudlinuxv1.CloudManaged)
var CloudManagedBackups = make(map[string]*cloudlinuxv1.CloudManagedBackup)
var CloudManagedRestores = make(map[string]*cloudlinuxv1.CloudManagedRestore)

// Implements prometheus.Collector
type CloudManagedCollector struct {
}

func (c CloudManagedCollector) Describe(ch chan<- *prometheus.Desc) {
	descriptors := []*prometheus.Desc{
		cmReady,
		cmReplicas,
		cmMemLimit,
		cmCPULimit,
		cmBackupSuccess,
		cmBackupStatus,
		cmRestoreSuccess,
		cmRestoreStatus,
	}
	for _, desc := range descriptors {
		ch <- desc
	}
}

func (c CloudManagedCollector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range CloudManageds {
		ch <- prometheus.MustNewConstMetric(
			cmReady, prometheus.GaugeValue, calcStatus(c),
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
	for _, c := range CloudManagedBackups {
		ch <- prometheus.MustNewConstMetric(
			cmBackupSuccess, prometheus.GaugeValue, calcStatus(c),
			c.Name, c.Namespace, c.Spec.ClusterName)
		ch <- prometheus.MustNewConstMetric(
			cmBackupStatus, prometheus.GaugeValue, 1,
			c.Name, c.Namespace, c.Spec.ClusterName)
	}
	for _, c := range CloudManagedRestores {
		ch <- prometheus.MustNewConstMetric(
			cmRestoreSuccess, prometheus.GaugeValue, calcStatus(c),
			c.Name, c.Namespace, c.Spec.ClusterName)
		ch <- prometheus.MustNewConstMetric(
			cmRestoreStatus, prometheus.GaugeValue, 1,
			c.Name, c.Namespace, c.Spec.ClusterName)
	}
}

func calcStatus(cmb interface{}) float64 {
	switch val := cmb.(type) {
	case *cloudlinuxv1.CloudManagedRestore:
		switch val.Status.Status {
		case cloudlinuxv1.BackupSuccessStatus:
			return 1
		default:
			return 0
		}
	case *cloudlinuxv1.CloudManagedBackup:
		switch val.Status.Status {
		case cloudlinuxv1.BackupSuccessStatus:
			return 1
		default:
			return 0
		}
	case *cloudlinuxv1.CloudManaged:
		switch val.Status.Status {
		case cloudlinuxv1.ClusterOkStatus:
			return 1
		default:
			return 0
		}
	default:
		panic(fmt.Sprintf("%s (%T) is not implemented", cmb, cmb))
	}
}

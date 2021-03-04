package monitoring

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/prometheus/client_golang/prometheus"
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
		"kuberlogic_ready",
		"KuberLogicServices application ready status",
		labels,
		nil)
	cmReplicas = prometheus.NewDesc(
		"kuberlogic_replicas",
		"KuberLogicServices application replicas",
		labels,
		nil)
	cmMemLimit = prometheus.NewDesc(
		"kuberlogic_memory_limit_bytes",
		"KuberLogicServices application memory limit",
		labels,
		nil)
	cmCPULimit = prometheus.NewDesc(
		"kuberlogic_cpu_limit_milliseconds",
		"KuberLogicServices application cpu limit",
		labels,
		nil)

	// Backup
	cmBackupSuccess = prometheus.NewDesc(
		"kuberlogicbackupschedule_success",
		"KuberLogicServices backup success",
		backupLabels,
		nil)
	cmBackupStatus = prometheus.NewDesc(
		"kuberlogicbackupschedule_status",
		"KuberLogicServices backup status",
		backupLabels,
		nil)
	// Restore
	cmRestoreSuccess = prometheus.NewDesc(
		"kuberlogicbackuprestore_success",
		"KuberLogicServices restore success",
		backupLabels,
		nil)
	cmRestoreStatus = prometheus.NewDesc(
		"kuberlogicbackuprestore_status",
		"KuberLogicServices backup's restore status",
		backupLabels,
		nil)
)

var KuberLogicServices = make(map[string]*kuberlogicv1.KuberLogicService)
var KuberLogicBackupSchedules = make(map[string]*kuberlogicv1.KuberLogicBackupSchedule)
var KuberLogicBackupRestores = make(map[string]*kuberlogicv1.KuberLogicBackupRestore)

// Implements prometheus.Collector
type KuberLogicCollector struct{}

func (c KuberLogicCollector) Describe(ch chan<- *prometheus.Desc) {
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

func (c KuberLogicCollector) Collect(ch chan<- prometheus.Metric) {
	for _, c := range KuberLogicServices {
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
	for _, c := range KuberLogicBackupSchedules {
		ch <- prometheus.MustNewConstMetric(
			cmBackupSuccess, prometheus.GaugeValue, calcStatus(c),
			c.Name, c.Namespace, c.Spec.ClusterName)
		ch <- prometheus.MustNewConstMetric(
			cmBackupStatus, prometheus.GaugeValue, 1,
			c.Name, c.Namespace, c.Spec.ClusterName)
	}
	for _, c := range KuberLogicBackupRestores {
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
	case *kuberlogicv1.KuberLogicBackupRestore:
		switch val.Status.Status {
		case kuberlogicv1.BackupSuccessStatus:
			return 1
		default:
			return 0
		}
	case *kuberlogicv1.KuberLogicBackupSchedule:
		switch val.Status.Status {
		case kuberlogicv1.BackupSuccessStatus:
			return 1
		default:
			return 0
		}
	case *kuberlogicv1.KuberLogicService:
		switch val.Status.Status {
		case kuberlogicv1.ClusterOkStatus:
			return 1
		default:
			return 0
		}
	default:
		panic(fmt.Sprintf("%s (%T) is not implemented", cmb, cmb))
	}
}

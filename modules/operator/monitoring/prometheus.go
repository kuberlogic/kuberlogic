/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package monitoring

import (
	"fmt"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
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

	// BaseBackup
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

// KuberLogicCollector implements prometheus.Collector and stores pointers to monitored resources
type KuberLogicCollector struct {
	mu  sync.RWMutex
	kls map[string]*kuberlogicv1.KuberLogicService
	klb map[string]*kuberlogicv1.KuberLogicBackupSchedule
	klr map[string]*kuberlogicv1.KuberLogicBackupRestore
}

func (c *KuberLogicCollector) MonitorKuberlogicService(kls *kuberlogicv1.KuberLogicService) error {
	return c.monitorResource(kls.Name+kls.Namespace, kls)
}

func (c *KuberLogicCollector) ForgetKuberlogicService(kls *kuberlogicv1.KuberLogicService) {
	delete(c.kls, kls.Name+kls.Namespace)
}

func (c *KuberLogicCollector) MonitorKuberlogicBackup(key string, klb *kuberlogicv1.KuberLogicBackupSchedule) error {
	return c.monitorResource(key, klb)
}

func (c *KuberLogicCollector) ForgetKuberlogicBackup(key string) {
	delete(c.klb, key)
}

func (c *KuberLogicCollector) MonitorKuberlogicRestore(key string, klr *kuberlogicv1.KuberLogicBackupRestore) error {
	return c.monitorResource(key, klr)
}

func (c *KuberLogicCollector) ForgetKuberlogicRestore(key string) {
	delete(c.klr, key)
}

func (c *KuberLogicCollector) Describe(ch chan<- *prometheus.Desc) {
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

func (c *KuberLogicCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, res := range c.kls {
		ch <- prometheus.MustNewConstMetric(
			cmReady, prometheus.GaugeValue, calcStatus(res),
			res.Name, res.Namespace, res.Spec.Type)
		ch <- prometheus.MustNewConstMetric(
			cmReplicas, prometheus.GaugeValue, float64(res.Spec.Replicas),
			res.Name, res.Namespace, res.Spec.Type)
		ch <- prometheus.MustNewConstMetric(
			cmMemLimit, prometheus.GaugeValue, float64(res.Spec.Resources.Limits.Memory().Value()),
			res.Name, res.Namespace, res.Spec.Type)
		ch <- prometheus.MustNewConstMetric(
			cmCPULimit, prometheus.GaugeValue, float64(res.Spec.Resources.Limits.Cpu().MilliValue()),
			res.Name, res.Namespace, res.Spec.Type)
	}
	for _, res := range c.klb {
		ch <- prometheus.MustNewConstMetric(
			cmBackupSuccess, prometheus.GaugeValue, calcStatus(res),
			res.Name, res.Namespace, res.Spec.ClusterName)
		ch <- prometheus.MustNewConstMetric(
			cmBackupStatus, prometheus.GaugeValue, 1,
			res.Name, res.Namespace, res.Spec.ClusterName)
	}
	for _, res := range c.klr {
		ch <- prometheus.MustNewConstMetric(
			cmRestoreSuccess, prometheus.GaugeValue, calcStatus(res),
			res.Name, res.Namespace, res.Spec.ClusterName)
		ch <- prometheus.MustNewConstMetric(
			cmRestoreStatus, prometheus.GaugeValue, 1,
			res.Name, res.Namespace, res.Spec.ClusterName)
	}

}

func (c *KuberLogicCollector) monitorResource(key string, res interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch val := res.(type) {
	case *kuberlogicv1.KuberLogicService:
		c.kls[key] = val
	case *kuberlogicv1.KuberLogicBackupSchedule:
		c.klb[key] = val
	case *kuberlogicv1.KuberLogicBackupRestore:
		c.klr[key] = val
	default:
		return fmt.Errorf("unknown resource type")
	}
	return nil
}

func NewCollector() *KuberLogicCollector {
	return &KuberLogicCollector{
		mu:  sync.RWMutex{},
		kls: make(map[string]*kuberlogicv1.KuberLogicService),
		klb: make(map[string]*kuberlogicv1.KuberLogicBackupSchedule),
		klr: make(map[string]*kuberlogicv1.KuberLogicBackupRestore),
	}
}

func calcStatus(cmb interface{}) float64 {
	switch val := cmb.(type) {
	case *kuberlogicv1.KuberLogicBackupRestore:
		switch val.IsSuccessful() {
		case true:
			return 1
		default:
			return 0
		}
	case *kuberlogicv1.KuberLogicBackupSchedule:
		switch val.IsSuccessful() {
		case true:
			return 1
		default:
			return 0
		}
	case *kuberlogicv1.KuberLogicService:
		status, _ := val.IsReady()
		switch status {
		case true:
			return 1
		default:
			return 0
		}
	default:
		panic(fmt.Sprintf("%s (%T) is not implemented", cmb, cmb))
	}
}

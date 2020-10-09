package monitoring

type MetricsMetadata struct {
	Name      string
	Namespace string
	Type      string
}

type MetricsStore struct {
	Meta *MetricsMetadata

	Ready    bool
	Replicas int32
	CPULimit int64
	MemLimit int64
}

const (
	nameLabel      = "name"
	namespaceLabel = "namespace"
	typeLabel      = "type"
)

var (
	labelList = []string{
		nameLabel,
		namespaceLabel,
		typeLabel,
	}
)

func (s *MetricsStore) Expose() {
	labels := map[string]string{
		nameLabel:      s.Meta.Name,
		namespaceLabel: s.Meta.Namespace,
		typeLabel:      s.Meta.Type,
	}

	exposeReadinessMetric(s.Ready, labels)
	exposeReplicasMetric(s.Replicas, labels)
	exposeCPULimitsMetric(s.CPULimit, labels)
	exposeMemLimitMetric(s.MemLimit, labels)
}

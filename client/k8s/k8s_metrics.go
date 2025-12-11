package k8s

import (
	"context"
	"fmt"

	"github.com/hosgf/element/types"
	"github.com/hosgf/element/uerrors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Metric struct {
	Namespace string                                          `json:"namespace,omitempty"`
	Name      string                                          `json:"name,omitempty"`
	Items     map[string]map[types.ResourceType]MetricDetails `json:"items,omitempty"`
}

type MetricDetails struct {
	Unit  string `json:"unit,omitempty"`
	Usage int64  `json:"usage,omitempty"`
}

type metricsOperation struct {
	*options
}

func (o *metricsOperation) List(ctx context.Context, namespace string) ([]*Metric, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{}
	data, err := o.metricsApi.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts)
	if err != nil {
		return nil, uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("获取Pod指标列表: namespace=%s", namespace))
	}
	metrics := make([]*Metric, 0, len(data.Items))
	for _, item := range data.Items {
		metric := &Metric{
			Namespace: item.GetNamespace(),
			Name:      item.GetName(),
			Items:     make(map[string]map[types.ResourceType]MetricDetails),
		}
		for _, c := range item.Containers {
			resources := make(map[types.ResourceType]MetricDetails)
			for name, quantity := range c.Usage {
				value, unit := types.Parse(quantity.String())
				switch name {
				case corev1.ResourceCPU:
					resources[types.ResourceCPU] = MetricDetails{
						Unit:  types.DefaultCpuUnit,
						Usage: types.FormatCpu(value, unit),
					}
				case corev1.ResourceMemory:
					resources[types.ResourceMemory] = MetricDetails{
						Unit:  types.DefaultMemoryUnit,
						Usage: types.FormatMemory(value, unit),
					}
				}
			}
			metric.Items[c.Name] = resources
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

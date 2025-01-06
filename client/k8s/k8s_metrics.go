package k8s

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/model/resource"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Metric struct {
	Namespace string                                             `json:"namespace,omitempty"`
	Name      string                                             `json:"name,omitempty"`
	Items     map[string]map[types.ResourceType]resource.Details `json:"items,omitempty"`
}

type metricsOperation struct {
	*options
}

func (o *metricsOperation) List(ctx context.Context, namespace string) ([]Metric, error) {
	if o.err != nil {
		return nil, o.err
	}
	opts := v1.ListOptions{}
	data, err := o.metricsApi.MetricsV1beta1().PodMetricses(namespace).List(ctx, opts)
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get Pod Metricses: %v", err)
	}
	metrics := make([]Metric, 0, len(data.Items))
	for _, item := range data.Items {
		metric := Metric{
			Namespace: item.GetNamespace(),
			Name:      item.GetName(),
			Items:     make(map[string]map[types.ResourceType]resource.Details),
		}
		for _, c := range item.Containers {
			resources := make(map[types.ResourceType]resource.Details)
			for name, quantity := range c.Usage {
				value, unit := types.Parse(quantity.String())
				switch name {
				case corev1.ResourceCPU:
					resources[types.ResourceCPU] = resource.Details{
						Type:  types.ResourceCPU,
						Unit:  unit,
						Usage: types.FormatCpu(value, unit),
					}
				case corev1.ResourceMemory:
					resources[types.ResourceMemory] = resource.Details{
						Type:  types.ResourceMemory,
						Unit:  unit,
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

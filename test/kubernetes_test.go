package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
	"github.com/hosgf/element/model/process"
	"github.com/hosgf/element/types"
)

func client() *k8s.Kubernetes {
	kubernetes := k8s.New(true, false)
	kubernetes.Init("")
	return kubernetes
}

func TestNodeTop(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	nodes, err := kubernetes.Nodes().Top(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(nodes)
}

func TestNamespaces(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	namespaces, err := kubernetes.Namespace().List(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(namespaces)
}

func TestCreateNamespace(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	isOk, err := kubernetes.Namespace().Apply(ctx, "test21", "test21")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(isOk)
}

func TestMetrics(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Metrics().List(ctx, "kube-system")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestResourceList(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Resource().Get(ctx)
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestPodList(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Pod().List(ctx, "kube-system")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestCreatePod(t *testing.T) {
	ctx := context.Background()
	pod := &k8s.Pod{
		Model: k8s.Model{
			Namespace: "test21",
			Name:      "mysql",
			App:       "mysqlapp",
			Group:     "mysqlgroup",
			Owner:     "mysqlowner",
			Scope:     "mysqlscope",
		},
		Containers: []*k8s.Container{
			{
				Name:       "mysql-sql",
				Image:      "hub.youede.com/base/mysql:5.7.36-security-v1",
				PullPolicy: "",
				Command:    []string{},
				Args:       []string{},
				Ports: []process.Port{
					{
						Name:       "http",
						Protocol:   types.ProtocolTcp,
						TargetPort: 3306,
					},
				},
				Resource: []process.Resource{
					{
						Type:    types.ResourceCPU,
						Unit:    "m",
						Minimum: 1,
						Maximum: 1,
					},
					{
						Type:    types.ResourceMemory,
						Unit:    "Gi",
						Minimum: 2,
						Maximum: 4,
					},
				},
				Env: map[string]string{
					"MYSQL_USER":          "Yaosu#DB@2024#",
					"MYSQL_PASSWORD":      "YaoSu",
					"MYSQL_ROOT_PASSWORD": "Yaosu#DB@2024#",
				},
			},
		},
	}
	kubernetes := client()
	err := kubernetes.Pod().Apply(ctx, pod)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestParse(t *testing.T) {
	fmt.Println(types.Parse("16384Mi"))
	fmt.Println(types.Parse("16384"))
	fmt.Println(types.Parse("16384u"))
	fmt.Println(types.Parse(""))
}

package test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
	"github.com/hosgf/element/model/progress"
	"github.com/hosgf/element/types"
)

func TestProgressList(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Progress().List(ctx, "sandbox")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestProgressDelete(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	var (
		num       = "01"
		namespace = "sandbox"
	)
	err := kubernetes.Progress().Destroy(ctx, namespace, "data-sandbox-"+num)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProgressRunning(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	var (
		num       = "01"
		namespace = "sandbox"
	)
	config := &k8s.ProcessGroupConfig{
		Namespace:   namespace,
		GroupName:   "data-sandbox-" + num,
		AllowUpdate: true,
		Labels: types.Labels{
			App:   "data-sandbox-" + num,
			Owner: "match-data-platform",
			Scope: "datasandbox",
		},
		Storage: make([]k8s.Storage, 0),
		Process: make([]k8s.ProcessConfig, 0),
	}
	config.Process = append(config.Process, toProgress(namespace, num))
	//config.Process = append(config.Process, toProgress2())

	config.Storage = append(config.Storage, toStorage())
	err := kubernetes.Progress().Running(ctx, config)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func toStorage() k8s.Storage {
	return k8s.Storage{
		Name:       "sandbox-storage",
		Type:       "pvc",
		AccessMode: types.ReadWriteOnce,
		Size:       "2Gi",
		Item:       "sandboxClaim",
	}
}

func toProgress(namespace, num string) k8s.ProcessConfig {
	return k8s.ProcessConfig{
		Name:        "data-sandbox-" + num,
		Service:     "data-sandbox-" + num + "-service",
		ServiceType: "NodePort",
		Source:      "hub.dataos.top/data_match_platform/sandbox-data:arm64v8-v1.0.0",
		Ports: []progress.Port{
			{
				Name:       "core",
				Protocol:   types.ProtocolTcp,
				Port:       28001,
				TargetPort: 28001,
			},
			{
				Name:       "datalab",
				Protocol:   types.ProtocolTcp,
				Port:       8888,
				TargetPort: 8888,
			},
		},
		Resource: []progress.Resource{
			{
				Type:    types.ResourceCPU,
				Unit:    "m",
				Minimum: 100,
			},
			{
				Type:    types.ResourceMemory,
				Unit:    "Mi",
				Minimum: 50,
			},
		},
		Env: []types.Environment{
			{
				Items: map[string]string{
					"APP_TYPE":         "data",
					"RUNTIME_ENV":      namespace,
					"REGISTER_ADDRESS": "data-platform-sandbox.sjchbigdata.svc.cluster.local:3099",
				},
			},
		},
		//Mounts: []k8s.Mount{
		//	{
		//		Name:    "sandboxStorage",
		//		Path:    "pvc",
		//		SubPath: "sandboxClaim",
		//	},
		//	{
		//		Name:    "sandboxStorage1",
		//		Path:    "pvc",
		//		SubPath: "sandboxClaim",
		//	},
		//},
	}
}

func toProgress2() k8s.ProcessConfig {
	return k8s.ProcessConfig{
		Name:    "container-10-2",
		Service: "data-sandbox-01",
		Source:  "hub.dataos.top/new_dataos_deploy/driver-container:v4.2.0",
		Command: []string{"java -jar -Xms100M -Xmx500M -XX:+UseG1GC -javaagent:/driver-container-interim.jar /driver-container-interim.jar -Dfile.encoding=utf-8 --driver.uploadConfig={\\\"password\\\":\\\"dataos@123\\\",\\\"port\\\":\\\"21\\\",\\\"ip\\\":\\\"192.168.130.207\\\",\\\"username\\\":\\\"ftpuser\\\"} --driver.localAddress=driver-dm-10.local.svc.cluster.local --driver.reuse=false --driver.health.port=3099 --driver.env=local --driver.name=container-10-2 --driver.websocket.enable=false "},
		Ports: []progress.Port{
			{
				Name:       "http",
				Protocol:   types.ProtocolTcp,
				Port:       18003,
				TargetPort: 18003,
			},
			{
				Name:       "ws",
				Protocol:   types.ProtocolTcp,
				Port:       18004,
				TargetPort: 18004,
			},
		},
		Resource: []progress.Resource{
			{
				Type:    types.ResourceCPU,
				Unit:    "m",
				Minimum: 100,
			},
			{
				Type:    types.ResourceMemory,
				Unit:    "Mi",
				Minimum: 50,
			},
		},
		Env: []types.Environment{
			{
				Items: map[string]string{
					"CONTAINER_NAME": "container-10-2",
					"PORT":           "18003",
					"MANAGER_LINK":   "http://driver-manager-service:18090",
					"RUN_ENV":        "sandbox",
				},
			},
		},
	}
}

package test

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/hosgf/element/client/k8s"
	"github.com/hosgf/element/model/process"
	"github.com/hosgf/element/types"
)

func TestProcessList(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Process().List(ctx, "sandbox")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestProcessGet(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	datas, err := kubernetes.Process().List(ctx, "sandbox")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestProcessStart(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	config := toProcessGroupConfig()
	err := kubernetes.Process().Start(ctx, config)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProcessRestart(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	// kubectl exec -it -n sandbox  data-sandbox-01-6f97c86f55-5gt44 -c  data-sandbox-01 -- /bin/bash
	err := kubernetes.Process().Restart(ctx, "sandbox", "data-sandbox-2-76b769d4cd-n76lz", "sandbox-2")
	//err := kubernetes.Process().Restart(ctx, "sandbox", "data-sandbox-01", "data-sandbox-01")
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestPodExec(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	// kubectl exec -it -n sandbox  data-sandbox-01-6f97c86f55-5gt44 -c  data-sandbox-01 -- /bin/bash
	// kubectl exec -it data-sandbox-2-76b769d4cd-n76lz -c sandbox-2 -n sandbox  -- bash /data/restart.sh
	data, err := kubernetes.Pod().Exec(ctx, "sandbox", "data-sandbox-2-76b769d4cd-6ph5x", "sandbox-3", "sh", "restart.sh")
	//err := kubernetes.Process().Restart(ctx, "sandbox", "data-sandbox-01", "data-sandbox-01")
	println(data)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProcessRestartGroup(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	err := kubernetes.Process().RestartGroup(ctx, "sandbox", "data-sandbox-01")
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProcessRestartApp(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	err := kubernetes.Process().RestartApp(ctx, "sandbox", "data-sandbox-01")
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProcessStop(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	var (
		num       = "01"
		namespace = "sandbox"
	)
	err := kubernetes.Process().Stop(ctx, namespace, "data-sandbox-"+num)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProcessDestroy(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	var (
		num       = "01"
		namespace = "sandbox"
	)
	err := kubernetes.Process().Destroy(ctx, namespace, "data-sandbox-"+num)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestProcessRunning(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	config := toProcessGroupConfig()
	err := kubernetes.Process().Running(ctx, config)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func toProcessGroupConfig() *k8s.ProcessGroupConfig {
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
		//Storage: make([]types.Storage, 0),
		Process: make([]k8s.ProcessConfig, 0),
	}
	//config.Storage = append(config.Storage, toStorage())

	config.Process = append(config.Process, toProcess(namespace, num))
	//config.Process = append(config.Process, toProcess2())
	return config
}

func toStorage() types.Storage {
	return types.Storage{
		Name:       "sandbox-storage1",
		Type:       "pvc",
		AccessMode: types.ReadWriteOnce,
		Size:       "2Gi",
		Path:       "/data",
		Item:       "ceph-rbd",
		//Resource: types.StorageResource{
		//	Type: "",
		//	Item: "",
		//},
	}
}

func toProcess(namespace, num string) k8s.ProcessConfig {
	return k8s.ProcessConfig{
		Name:        "data-sandbox-" + num,
		Service:     "data-sandbox-" + num + "-service",
		ServiceType: "NodePort",
		Source:      "hub.dataos.top/data_match_platform/sandbox-data:arm64v8-v1.0.0",
		Ports: []process.Port{
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
		Resource: []process.Resource{
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
		//Mounts: []types.Mount{
		//	{
		//		Name: "sandbox-storage",
		//		Path: "/data/sandbox",
		//		//SubPath: "sandboxClaim",
		//	},
		//	//{
		//	//	Name:    "sandboxStorage1",
		//	//	Path:    "pvc",
		//	//	SubPath: "sandboxClaim",
		//	//},
		//},
	}
}

func toProcess2() k8s.ProcessConfig {
	return k8s.ProcessConfig{
		Name:    "container-10-2",
		Service: "data-sandbox-01",
		Source:  "hub.dataos.top/new_dataos_deploy/driver-container:v4.2.0",
		Command: []string{"java -jar -Xms100M -Xmx500M -XX:+UseG1GC -javaagent:/driver-container-interim.jar /driver-container-interim.jar -Dfile.encoding=utf-8 --driver.uploadConfig={\\\"password\\\":\\\"dataos@123\\\",\\\"port\\\":\\\"21\\\",\\\"ip\\\":\\\"192.168.130.207\\\",\\\"username\\\":\\\"ftpuser\\\"} --driver.localAddress=driver-dm-10.local.svc.cluster.local --driver.reuse=false --driver.health.port=3099 --driver.env=local --driver.name=container-10-2 --driver.websocket.enable=false "},
		Ports: []process.Port{
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
		Resource: []process.Resource{
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

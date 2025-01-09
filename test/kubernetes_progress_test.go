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
	datas, err := kubernetes.Progress().List(ctx, "local")
	if err != nil {
		t.Fatal(err)
		return
	}
	g.Dump(datas)
}

func TestDeleteProgress(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	err := kubernetes.Progress().Destroy(ctx, "local", "driver-dm-10")
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func TestRunningProgress(t *testing.T) {
	ctx := context.Background()
	kubernetes := client()
	config := &k8s.ProcessGroupConfig{
		Namespace:   "local",
		GroupName:   "driver-dm-10",
		AllowUpdate: false,
		Labels: types.Labels{
			App:   "driver-container-10",
			Owner: "driver-manage",
			Scope: "driver-container",
		},
		Process: make([]k8s.ProcessConfig, 0),
	}
	config.Process = append(config.Process, toProgress1())
	config.Process = append(config.Process, toProgress2())
	err := kubernetes.Progress().Running(ctx, config)
	if err != nil {
		t.Fatal(err)
		return
	}
	println("--------------------------------------------")
}

func toProgress1() k8s.ProcessConfig {
	return k8s.ProcessConfig{
		Name:    "container-10-1",
		Service: "driver-dm-10",
		Source:  "hub.dataos.top/new_dataos_deploy/driver-container:v4.2.0",
		Command: []string{"java -jar -Xms100M -Xmx500M -XX:+UseG1GC -javaagent:/driver-container-interim.jar /driver-container-interim.jar -Dfile.encoding=utf-8 --driver.uploadConfig={\\\"password\\\":\\\"dataos@123\\\",\\\"port\\\":\\\"21\\\",\\\"ip\\\":\\\"192.168.130.207\\\",\\\"username\\\":\\\"ftpuser\\\"} --driver.localAddress=driver-dm-10.local.svc.cluster.local --driver.reuse=false --driver.health.port=3099 --driver.env=local --driver.name=container-10-1 --driver.websocket.enable=false "},
		Ports: []progress.Port{
			{
				Name:       "http",
				Protocol:   "tcp",
				Port:       18000,
				TargetPort: 18000,
			},
			{
				Name:       "ws",
				Protocol:   types.ProtocolTcp,
				Port:       18001,
				TargetPort: 18001,
			},
		},
		Resource: []progress.Resource{
			{
				Type:    types.ResourceCPU,
				Unit:    "m",
				Minimum: 100,
				Maximum: 200,
			},
			{
				Type:    types.ResourceMemory,
				Unit:    "Mi",
				Minimum: 50,
				Maximum: 2048,
			},
		},
		Env: []types.Environment{
			{
				Items: map[string]string{
					"CONTAINER_NAME": "container-10-1",
					"PORT":           "18000",
					"MANAGER_LINK":   "http://driver-manager-service:18090",
					"RUN_ENV":        "local",
				},
			},
		},
	}
}

func toProgress2() k8s.ProcessConfig {
	return k8s.ProcessConfig{
		Name:    "container-10-2",
		Service: "driver-dm-10",
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
				Maximum: 200,
			},
			{
				Type:    types.ResourceMemory,
				Unit:    "Mi",
				Minimum: 50,
				Maximum: 2048,
			},
		},
		Env: []types.Environment{
			{
				Items: map[string]string{
					"CONTAINER_NAME": "container-10-2",
					"PORT":           "18003",
					"MANAGER_LINK":   "http://driver-manager-service:18090",
					"RUN_ENV":        "local",
				},
			},
		},
	}
}

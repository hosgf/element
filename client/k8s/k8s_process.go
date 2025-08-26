package k8s

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/process"
)

type processOperation struct {
	*options
	k8s *Kubernetes
}

func (o *processOperation) List(ctx context.Context, namespace string) ([]*process.Process, error) {
	if o.err != nil {
		return nil, o.err
	}
	var (
		list    = make([]*process.Process, 0)
		metrics = make(map[string]*Metric)
		svcs    = make(map[string][]*Service)
		now     = gtime.Now().Timestamp()
	)
	// 获取SVC
	services, err := o.k8s.Service().List(ctx, namespace)
	if err != nil {
		logger.Warningf(ctx, "---> SVC信息采集失败 err: %+v \r\n", err.Error())
	} else {
		if o.isDebug {
			logger.Debugf(ctx, "---> SVC信息采集成功 size %d \r\n", len(services))
		}
		for _, s := range services {
			svc, ok := svcs[s.Group]
			if !ok {
				svc = make([]*Service, 0)
			}
			svc = append(svc, s)
			svcs[s.Group] = svc
		}
	}
	// 获取POD
	pods, err := o.k8s.Pod().List(ctx, namespace)
	if err != nil {
		return list, err
	}
	if o.isDebug {
		logger.Debugf(ctx, "---> POD信息采集成功 size: %d \r\n", len(pods))
	}
	if len(pods) < 1 {
		return list, nil
	}
	// 获取pod资源
	ms, err := o.k8s.Metrics().List(ctx, namespace)
	if err != nil {
		logger.Warningf(ctx, "---> POD资源信息采集失败 err: %+v \r\n", err.Error())
	} else {
		if o.isDebug {
			logger.Debugf(ctx, "---> POD资源信息采集成功 %+v \r\n", ms)
		}
		for _, m := range ms {
			metrics[m.Name] = m
		}
	}
	for _, pod := range pods {
		ps := pod.ToProcess(svcs[pod.Group], metrics[pod.Name], now)
		list = append(list, ps...)
	}
	return list, nil
}

func (o *processOperation) Running(ctx context.Context, config *ProcessGroupConfig) error {
	if o.err != nil {
		return o.err
	}
	if _, err := o.k8s.Namespace().Apply(ctx, config.Namespace, config.Labels.Owner); err != nil {
		return err
	}

	svcs := config.toServices()
	for _, s := range svcs {
		if err := o.k8s.Service().Apply(ctx, s); err != nil {
			return err
		}
	}
	pod := config.toPod()
	if pod == nil {
		return nil
	}
	if err := o.k8s.StorageResource().BatchApply(ctx, config.toModel(), config.Storage); err != nil {
		return err
	}
	if err := o.k8s.Storage().BatchApply(ctx, config.toModel(), config.Storage); err != nil {
		return err
	}
	if err := o.k8s.Pod().Apply(ctx, pod); err != nil {
		return err
	}
	return nil
}

// ensureStorage 确保 PV/PVC 存在，缺失则批量创建
func (o *processOperation) ensureStorage(ctx context.Context, config *ProcessGroupConfig) error {
	if len(config.Storage) == 0 {
		return nil
	}
	fmt.Printf("[ensureStorage] start, namespace=%s, items=%d\n", config.Namespace, len(config.Storage))
	needResApply := false
	needPvcApply := false
	for _, s := range config.Storage {
		if len(s.Name) < 1 {
			continue
		}
		// PV 检查：不存在或处于删除中
		if pv, err := o.k8s.api.CoreV1().PersistentVolumes().Get(ctx, s.Name, v1.GetOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				fmt.Printf("[ensureStorage] PV not found, name=%s -> needResApply\n", s.Name)
				needResApply = true
			} else {
				return err
			}
		} else if pv != nil && pv.DeletionTimestamp != nil {
			fmt.Printf("[ensureStorage] PV deleting, wait to disappear, name=%s\n", s.Name)
			deadline := time.Now().Add(60 * time.Second)
			for {
				_, err := o.k8s.api.CoreV1().PersistentVolumes().Get(ctx, s.Name, v1.GetOptions{})
				if apierrors.IsNotFound(err) {
					needResApply = true
					break
				}
				if err != nil {
					return err
				}
				if time.Now().After(deadline) {
					return gerror.NewCodef(gcode.CodeNotImplemented, "等待PV删除超时")
				}
				time.Sleep(2 * time.Second)
			}
		}

		// PVC 检查：不存在或处于删除中
		if pvc, err := o.k8s.api.CoreV1().PersistentVolumeClaims(config.Namespace).Get(ctx, s.Name, v1.GetOptions{}); err != nil {
			if apierrors.IsNotFound(err) {
				fmt.Printf("[ensureStorage] PVC not found, ns=%s name=%s -> needPvcApply\n", config.Namespace, s.Name)
				needPvcApply = true
			} else {
				return err
			}
		} else if pvc != nil && pvc.DeletionTimestamp != nil {
			fmt.Printf("[ensureStorage] PVC deleting, wait to disappear, ns=%s name=%s\n", config.Namespace, s.Name)
			deadline := time.Now().Add(60 * time.Second)
			for {
				_, err := o.k8s.api.CoreV1().PersistentVolumeClaims(config.Namespace).Get(ctx, s.Name, v1.GetOptions{})
				if apierrors.IsNotFound(err) {
					needPvcApply = true
					break
				}
				if err != nil {
					return err
				}
				if time.Now().After(deadline) {
					return gerror.NewCodef(gcode.CodeNotImplemented, "等待PVC删除超时")
				}
				time.Sleep(2 * time.Second)
			}
		}
	}
	if needResApply {
		fmt.Printf("[ensureStorage] applying PV batch, items=%d\n", len(config.Storage))
		if err := o.k8s.StorageResource().BatchApply(ctx, config.toModel(), config.Storage); err != nil {
			return err
		}
	}
	if needPvcApply {
		fmt.Printf("[ensureStorage] applying PVC batch, ns=%s items=%d\n", config.Namespace, len(config.Storage))
		if err := o.k8s.Storage().BatchApply(ctx, config.toModel(), config.Storage); err != nil {
			return err
		}
	}
	// 等待所有目标 PVC 可用（存在、非删除中、Bound）
	if len(config.Storage) < 1 {
		return nil
	}
	deadline := time.Now().Add(60 * time.Second)
	for {
		allReady := true
		for _, s := range config.Storage {
			if len(s.Name) < 1 {
				continue
			}
			pvc, err := o.k8s.api.CoreV1().PersistentVolumeClaims(config.Namespace).Get(ctx, s.Name, v1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					allReady = false
					fmt.Printf("[ensureStorage] waiting PVC appear, ns=%s name=%s\n", config.Namespace, s.Name)
					break
				}
				return err
			}
			if pvc.DeletionTimestamp != nil {
				allReady = false
				fmt.Printf("[ensureStorage] waiting PVC deletion finish, ns=%s name=%s\n", config.Namespace, s.Name)
				break
			}
			if string(pvc.Status.Phase) != "Bound" {
				allReady = false
				fmt.Printf("[ensureStorage] waiting PVC bound, ns=%s name=%s phase=%s\n", config.Namespace, s.Name, string(pvc.Status.Phase))
				break
			}
		}
		if allReady {
			fmt.Printf("[ensureStorage] all PVC ready, ns=%s\n", config.Namespace)
			break
		}
		if time.Now().After(deadline) {
			fmt.Printf("[ensureStorage] wait PVC ready timeout, ns=%s\n", config.Namespace)
			return gerror.NewCodef(gcode.CodeNotImplemented, "等待PVC就绪超时")
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (o *processOperation) Start(ctx context.Context, config *ProcessGroupConfig) error {
	if o.err != nil {
		return o.err
	}
	pod := config.toPod()
	if pod == nil {
		return nil
	}
	// 仅在缺失时创建 存储资源(PV) 与 存储(PVC)，避免 PVC 不存在导致调度失败
	if err := o.ensureStorage(ctx, config); err != nil {
		return err
	}
	if err := o.k8s.Pod().Apply(ctx, pod); err != nil {
		return err
	}
	return nil
}

func (o *processOperation) Stop(ctx context.Context, namespace string, groups ...string) error {
	if o.err != nil {
		return o.err
	}
	if groups == nil || len(groups) < 1 {
		return gerror.NewCodef(gcode.CodeNotImplemented, "请传入要删除的进程组名称")
	}
	if err := o.k8s.Pod().DeleteGroup(ctx, namespace, groups...); err != nil {
		return err
	}
	return nil
}

func (o *processOperation) Destroy(ctx context.Context, namespace string, groups ...string) error {
	if o.err != nil {
		return o.err
	}
	if groups == nil || len(groups) < 1 {
		return gerror.NewCodef(gcode.CodeNotImplemented, "请传入要删除的进程组名称")
	}
	if err := o.k8s.Service().DeleteGroup(ctx, namespace, groups...); err != nil {
		return err
	}
	if err := o.k8s.Pod().DeleteGroup(ctx, namespace, groups...); err != nil {
		return err
	}
	if err := o.k8s.Storage().DeleteByGroup(ctx, true, namespace, groups...); err != nil {
		return err
	}
	return nil
}

func (o *processOperation) Restart(ctx context.Context, namespace, group, process string, cmd ...string) error {
	cmds := append([]string{"/bin/bash", "-c", "restart.sh"}, cmd...)
	//cmds := append([]string{"/bin/bash", "-c"}, cmd...)
	_, err := o.k8s.Pod().Command(ctx, namespace, group, process, cmds...)
	return err
}

func (o *processOperation) RestartGroup(ctx context.Context, namespace, group string) error {
	return o.k8s.Pod().RestartGroup(ctx, namespace, group)
}

func (o *processOperation) RestartApp(ctx context.Context, namespace, group string) error {
	return o.k8s.Pod().RestartApp(ctx, namespace, group)
}

func (o *processOperation) Command(ctx context.Context, namespace, group, process string, cmd ...string) (string, error) {
	return o.k8s.Pod().Command(ctx, namespace, group, process, cmd...)
}

func (o *processOperation) Logger(ctx context.Context, namespace, group, process string, config ProcessLogger) (io.ReadCloser, error) {
	return o.k8s.Pod().Logger(ctx, namespace, group, process, config)
}

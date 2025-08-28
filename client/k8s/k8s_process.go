package k8s

import (
	"context"
	"io"
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/process"
	"github.com/hosgf/element/types"
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

func (o *processOperation) Start(ctx context.Context, config *ProcessGroupConfig) error {
	if o.err != nil {
		return o.err
	}
	logger.Info(ctx, "[Start] begin, ns=", config.Namespace, ", group=", config.GroupName, ", replicas=", config.Replicas)
	pod := config.toPod()
	if pod == nil {
		logger.Warningf(ctx, "[Start] toPod returned nil, skip. ns=%s group=%s", config.Namespace, config.GroupName)
		return nil
	}
	// 打印存储配置
	logger.Debugf(ctx, "[Start] storage.len=%d\n", len(config.Storage))
	if len(config.Storage) > 0 {
		names := make([]string, 0, len(config.Storage))
		for _, s := range config.Storage {
			names = append(names, s.Name)
		}
		logger.Debugf(ctx, "[Start] storage.names=%v", names)
	}

	// 确保 SVC 存在，缺失则创建
	if err := o.ensureServices(ctx, config); err != nil {
		logger.Warningf(ctx, "[Start] ensureServices failed: %v", err)
		return err
	}
	logger.Debugf(ctx, "[Start] ensureServices done, ns=%s", config.Namespace)

	// 确保 Storage 存在，缺失则创建
	if err := o.ensureStorage(ctx, config); err != nil {
		logger.Warningf(ctx, "[Start] ensureStorage failed: %v", err)
		return err
	}
	logger.Debugf(ctx, "[Start] ensureStorage done, ns=%s", config.Namespace)
	// 额外校验：确保即将引用的 PVC 均已存在，避免命名不一致导致的调度失败
	if pod.Storage != nil && len(pod.Storage) > 0 {
		for _, s := range pod.Storage {
			if s.ToStorageType().String() != "pvc" {
				continue
			}
			if len(s.Name) < 1 {
				continue
			}
			if has, err := o.k8s.Storage().Exists(ctx, pod.Namespace, s.Name); err != nil {
				logger.Warningf(ctx, "[Start] check PVC exists error: ns=%s name=%s err=%v", pod.Namespace, s.Name, err)
				return err
			} else if !has {
				logger.Warningf(ctx, "[start] PVC missing before apply: ns=%s name=%s", pod.Namespace, s.Name)
				return gerror.NewCodef(gcode.CodeNotImplemented, "PVC缺失: %s", s.Name)
			}
		}
	}

	logger.Info(ctx, "[Start] applying Pod, ns=", pod.Namespace, ", group=", pod.Group)
	if err := o.k8s.Pod().Apply(ctx, pod); err != nil {
		logger.Warningf(ctx, "[Start] apply Pod failed: %v", err)
		return err
	}
	logger.Info(ctx, "[Start] done, ns=", pod.Namespace, ", group=", pod.Group)
	return nil
}

// ensureServices 确保 SVC 存在，缺失则批量创建
func (o *processOperation) ensureServices(ctx context.Context, config *ProcessGroupConfig) error {
	svcs := config.toServices()
	if len(svcs) == 0 {
		return nil
	}
	logger.Debugf(ctx, "[ensureServices] start, namespace=%s, items=%d", config.Namespace, len(svcs))

	for _, svc := range svcs {
		if len(svc.Name) < 1 {
			continue
		}

		// 检查 SVC 是否存在
		exists, err := o.k8s.Service().Exists(ctx, config.Namespace, svc.Name)
		if err != nil {
			return err
		}

		if !exists {
			logger.Debugf(ctx, "[ensureServices] Service not found, creating: ns=%s name=%s", config.Namespace, svc.Name)
			if err := o.k8s.Service().Apply(ctx, svc); err != nil {
				return err
			}
		} else {
			logger.Debugf(ctx, "[ensureServices] Service already exists: ns=%s name=%s", config.Namespace, svc.Name)
		}
	}

	logger.Debugf(ctx, "[ensureServices] done, namespace=%s", config.Namespace)
	return nil
}

// ensureStorage 确保 StorageResource(PV) 和 Storage(PVC) 存在，缺失则批量创建
func (o *processOperation) ensureStorage(ctx context.Context, config *ProcessGroupConfig) error {
	if len(config.Storage) == 0 {
		return nil
	}
	logger.Debugf(ctx, "[ensureStorage] start, namespace=%s, items=%d", config.Namespace, len(config.Storage))

	// 先处理 StorageResource (PV)
	if err := o.ensureStorageResources(ctx, config); err != nil {
		return err
	}
	// 再处理 Storage (PVC)
	if err := o.ensureStorages(ctx, config); err != nil {
		return err
	}
	// 最后等待所有 Storage (PVC) 就绪
	return o.waitStoragesBound(ctx, config.Namespace, config.Storage, 60*time.Second)
}

// ensureStorageResources 确保 StorageResource (PV) 存在，缺失则批量创建
func (o *processOperation) ensureStorageResources(ctx context.Context, config *ProcessGroupConfig) error {
	needStorageResourceApply := false
	for _, s := range config.Storage {
		if len(s.Name) < 1 {
			continue
		}

		// 检查 StorageResource (PV) 是否存在或正在删除
		exists, err := o.k8s.StorageResource().Exists(ctx, s.Name)
		if err != nil {
			return err
		}

		if !exists {
			logger.Debugf(ctx, "[ensureStorageResources] StorageResource not found, name=%s -> needStorageResourceApply", s.Name)
			needStorageResourceApply = true
			continue
		}

		// 如果存在，检查是否正在删除
		isDeleting, err := o.k8s.StorageResource().IsDeleting(ctx, s.Name)
		if err != nil {
			return err
		}

		if isDeleting {
			logger.Debugf(ctx, "[ensureStorageResources] StorageResource is being deleted, waiting for completion, name=%s", s.Name)
			if err := o.k8s.StorageResource().WaitDeleted(ctx, s.Name, 60*time.Second); err != nil {
				return err
			}
			needStorageResourceApply = true
		}
	}

	if needStorageResourceApply {
		logger.Debugf(ctx, "[ensureStorageResources] applying StorageResource batch, items=%d", len(config.Storage))
		if err := o.k8s.StorageResource().BatchApply(ctx, config.toModel(), config.Storage); err != nil {
			return err
		}
	}
	return nil
}

// ensureStorages 确保 Storage (PVC) 存在，缺失则批量创建
func (o *processOperation) ensureStorages(ctx context.Context, config *ProcessGroupConfig) error {
	needStorageApply := false
	for _, s := range config.Storage {
		if len(s.Name) < 1 {
			continue
		}

		// 检查 Storage (PVC) 是否存在或正在删除
		exists, err := o.k8s.Storage().Exists(ctx, config.Namespace, s.Name)
		if err != nil {
			return err
		}

		if !exists {
			logger.Debugf(ctx, "[ensureStorages] Storage not found, ns=%s name=%s -> needStorageApply", config.Namespace, s.Name)
			needStorageApply = true
			continue
		}

		// 如果存在，检查是否正在删除
		isDeleting, err := o.k8s.Storage().IsDeleting(ctx, config.Namespace, s.Name)
		if err != nil {
			return err
		}

		if isDeleting {
			logger.Debugf(ctx, "[ensureStorages] Storage is being deleted, waiting for completion, ns=%s name=%s", config.Namespace, s.Name)
			if err := o.k8s.Storage().WaitDeleted(ctx, config.Namespace, s.Name, 60*time.Second); err != nil {
				return err
			}
			needStorageApply = true
		}
	}

	if needStorageApply {
		logger.Debugf(ctx, "[ensureStorages] applying Storage batch, ns=%s items=%d", config.Namespace, len(config.Storage))
		if err := o.k8s.Storage().BatchApply(ctx, config.toModel(), config.Storage); err != nil {
			return err
		}
	}
	return nil
}

// waitStoragesBound 等待所有 Storage (PVC) 就绪
func (o *processOperation) waitStoragesBound(ctx context.Context, ns string, storages []types.Storage, timeout time.Duration) error {
	if len(storages) == 0 {
		return nil
	}

	deadline := time.Now().Add(timeout)
	for {
		allReady := true
		for _, s := range storages {
			if len(s.Name) < 1 {
				continue
			}

			// 使用 Storage 接口检查状态
			_, err := o.k8s.Storage().Get(ctx, ns, s.Name)
			if err != nil {
				if apierrors.IsNotFound(err) {
					allReady = false
					logger.Debugf(ctx, "[waitStoragesBound] waiting Storage appear, ns=%s name=%s", ns, s.Name)
					break
				}
				return err
			}
		}

		if allReady {
			logger.Debugf(ctx, "[waitStoragesBound] all Storage ready, ns=%s", ns)
			return nil
		}

		if time.Now().After(deadline) {
			logger.Warningf(ctx, "[waitStoragesBound] wait Storage ready timeout, ns=%s", ns)
			return gerror.NewCodef(gcode.CodeNotImplemented, "等待Storage就绪超时")
		}

		time.Sleep(2 * time.Second)
	}
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

package k8s

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/hosgf/element/logger"
	"github.com/hosgf/element/model/progress"
)

type progressOperation struct {
	*options
	k8s *Kubernetes
}

func (o *progressOperation) List(ctx context.Context, namespace string) ([]*progress.Progress, error) {
	if o.err != nil {
		return nil, o.err
	}
	var (
		list    = make([]*progress.Progress, 0)
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
			logger.Debugf(ctx, "---> SVC信息采集成功 %+v \r\n", services)
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
		logger.Debugf(ctx, "---> POD信息采集成功 size: %+v \r\n", pods)
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
		ps := pod.ToProgress(svcs[pod.Group], metrics[pod.Name], now)
		list = append(list, ps...)
	}
	return list, nil
}

func (o *progressOperation) Running(ctx context.Context, config *ProcessGroupConfig) error {
	if o.err != nil {
		return o.err
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
	if err := o.k8s.Pod().Apply(ctx, pod); err != nil {
		return err
	}
	return nil
}

func (o *progressOperation) Start(ctx context.Context, config *ProcessGroupConfig) error {
	return nil
}

func (o *progressOperation) Stop(ctx context.Context, config *ProcessGroupConfig) error {
	return nil
}

func (o *progressOperation) Destroy(ctx context.Context, namespace string, groups ...string) error {
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
	return nil
}

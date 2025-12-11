package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/types"
	"github.com/hosgf/element/uerrors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PersistentStorage struct {
	Model
	types.Storage
	TargetName string
}

func (s *PersistentStorage) toPvc() *corev1.PersistentVolumeClaim {
	pvcs := corev1.PersistentVolumeClaimSpec{
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(s.ToAccessMode())},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)},
			Limits:   corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)},
		},
	}
	if len(s.TargetName) > 0 {
		pvcs.VolumeName = s.TargetName
	}
	switch s.ToStorageType() {
	case types.StoragePVC:
		if s.Item != nil {
			item := gconv.String(s.Item)
			pvcs.StorageClassName = &item
		}
	}
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      s.Storage.Name,
			Namespace: s.Namespace,
			Labels:    s.labels(),
		},
		Spec: pvcs,
	}
}

func (s *PersistentStorage) updatePvc(data *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	data.Spec.Resources.Limits = corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)}
	return data
}

type storageOperation struct {
	*options
	k8s *Kubernetes
}

func (o *storageOperation) Get(ctx context.Context, namespace, name string) (*types.Storage, error) {
	if o.err != nil {
		return nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("获取Storage: namespace=%s, name=%s", namespace, name))
	}
	return o.toStorage(storage), nil
}

func (o *storageOperation) Exists(ctx context.Context, namespace, name string) (bool, error) {
	has, _, err := o.exists(ctx, namespace, name)
	return has, err
}

func (o *storageOperation) Apply(ctx context.Context, storage *PersistentStorage) error {
	if storage == nil {
		return nil
	}
	if o.err != nil {
		return o.err
	}
	if has, _, err := o.exists(ctx, storage.Namespace, storage.Storage.Name); has {
		if err != nil {
			return err
		}
		if storage.AllowUpdate {
			return nil
			//return o.update(ctx, data, storage)
		}
		return uerrors.NewBizLogicError(uerrors.CodeResourceConflict,
			fmt.Sprintf("Storage已存在: namespace=%s, name=%s, item=%v", storage.Namespace, storage.Storage.Name, storage.Item))
	}
	return o.create(ctx, storage)
}

func (o *storageOperation) BatchApply(ctx context.Context, model Model, storage []types.Storage) error {
	if storage == nil || len(storage) < 1 {
		return nil
	}
	model.AllowUpdate = true
	for _, s := range storage {
		ps := &PersistentStorage{
			Model:   model,
			Storage: s,
		}
		if err := o.Apply(ctx, ps); err != nil {
			return err
		}
	}
	return nil
}

func (o *storageOperation) Delete(ctx context.Context, delRes bool, namespace string, name ...string) error {
	if o.err != nil {
		return o.err
	}

	var lastErr error
	for _, n := range name {
		if has, _, err := o.exists(ctx, namespace, n); has {
			if err != nil {
				lastErr = err
				continue
			}
			if err := o.delete(ctx, namespace, n); err != nil {
				lastErr = err
			}
		}

		if delRes {
			if err := o.k8s.StorageResource().Delete(ctx, n); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

func (o *storageOperation) DeleteByGroup(ctx context.Context, delRes bool, namespace string, groups ...string) error {
	if o.err != nil {
		return o.err
	}

	var lastErr error
	for _, g := range groups {
		if has, list, err := o.existsByGroup(ctx, namespace, g); has {
			if err != nil {
				lastErr = err
				continue
			}
			if list != nil && len(list.Items) > 0 {
				for _, i := range list.Items {
					if err := o.delete(ctx, namespace, i.Name); err != nil {
						lastErr = err
					}
				}
			}
		}

		if delRes {
			if err := o.k8s.StorageResource().DeleteByGroup(ctx, g); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

func (o *storageOperation) create(ctx context.Context, storage *PersistentStorage) error {
	pvc := storage.toPvc()
	if o.isTest {
		return nil
	}
	_, err := o.api.CoreV1().PersistentVolumeClaims(storage.Namespace).Create(ctx, pvc, v1.CreateOptions{})
	if err != nil {
		return uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("创建Storage: namespace=%s, name=%s", storage.Namespace, storage.Storage.Name))
	}
	return nil
}

func (o *storageOperation) update(ctx context.Context, data *corev1.PersistentVolumeClaim, storage *PersistentStorage) error {
	_, err := o.api.CoreV1().PersistentVolumeClaims(storage.Namespace).Update(ctx, storage.updatePvc(data), v1.UpdateOptions{})
	if err != nil {
		return uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("更新Storage: namespace=%s, name=%s", storage.Namespace, storage.Storage.Name))
	}
	return nil
}

func (o *storageOperation) delete(ctx context.Context, namespace, name string) error {
	err := o.api.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		return uerrors.WrapKubernetesError(ctx, err, fmt.Sprintf("删除Storage: namespace=%s, name=%s", namespace, name))
	}
	return nil
}

func (o *storageOperation) WaitDeleted(ctx context.Context, namespace, name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		_, err := o.api.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, v1.GetOptions{})
		if errors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return uerrors.NewKubernetesError(ctx, "等待PVC删除", "超时",
				fmt.Sprintf("namespace=%s, name=%s, timeout=%v", namespace, name, timeout))
		}
		time.Sleep(2 * time.Second)
	}
}

func (o *storageOperation) IsDeleting(ctx context.Context, namespace, name string) (bool, error) {
	pvc, err := o.api.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return pvc.DeletionTimestamp != nil, nil
}

func (o *storageOperation) existsByGroup(ctx context.Context, namespace, group string) (bool, *corev1.PersistentVolumeClaimList, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumeClaims(namespace).List(ctx, v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", types.LabelGroup, group)})
	has, err := o.isExist(ctx, storage, err, fmt.Sprintf("检查Storage是否存在: namespace=%s, group=%s", namespace, group))
	return has, storage, err
}

func (o *storageOperation) exists(ctx context.Context, namespace, name string) (bool, *corev1.PersistentVolumeClaim, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, v1.GetOptions{})
	has, err := o.isExist(ctx, storage, err, fmt.Sprintf("检查Storage是否存在: namespace=%s, name=%s", namespace, name))
	return has, storage, err
}

func (o *storageOperation) toStorage(data *corev1.PersistentVolumeClaim) *types.Storage {
	if data == nil {
		return nil
	}
	sc := data.Spec
	s := &types.Storage{
		Name:       data.Name,
		AccessMode: types.AccessMode(sc.AccessModes[0]),
		Size:       sc.Resources.Limits.Storage().String(),
		Item:       sc.StorageClassName,
	}
	return s
}

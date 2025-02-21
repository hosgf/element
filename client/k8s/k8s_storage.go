package k8s

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/util/gconv"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PersistentStorage struct {
	Model
	types.Storage
}

func (s *PersistentStorage) toPvc() *corev1.PersistentVolumeClaim {
	pvcs := corev1.PersistentVolumeClaimSpec{
		VolumeName:  s.Storage.Name,
		AccessModes: []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(s.ToAccessMode())},
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)},
			Limits:   corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)},
		},
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
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get storages: %v", err)
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
		return gerror.NewCodef(gcode.CodeNotImplemented, "Storage: %s, Item: %v 已存在!", storage.Storage.Name, storage.Item)
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
	var err error
	for _, n := range name {
		if has, _, err := o.exists(ctx, namespace, n); has {
			if err != nil {
				return err
			}
			err = o.delete(ctx, namespace, n)
		}
		if delRes {
			err = o.k8s.StorageResource().Delete(ctx, n)
		}
	}
	return err
}

func (o *storageOperation) DeleteByGroup(ctx context.Context, delRes bool, namespace string, groups ...string) error {
	if o.err != nil {
		return o.err
	}
	var err error
	for _, g := range groups {
		if has, list, err := o.existsByGroup(ctx, namespace, g); has {
			if err != nil {
				return err
			}
			if list != nil && list.Size() > 0 {
				for _, i := range list.Items {
					err = o.delete(ctx, namespace, i.Name)
				}
			}
		}
		if delRes {
			err = o.k8s.StorageResource().DeleteByGroup(ctx, g)
		}
	}
	return err
}

func (o *storageOperation) create(ctx context.Context, storage *PersistentStorage) error {
	pvc := storage.toPvc()
	if o.isTest {
		return nil
	}
	_, err := o.api.CoreV1().PersistentVolumeClaims(storage.Namespace).Create(ctx, pvc, v1.CreateOptions{})
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to Create apps Storage: %v", err)
	}
	return nil
}

func (o *storageOperation) update(ctx context.Context, data *corev1.PersistentVolumeClaim, storage *PersistentStorage) error {
	_, err := o.api.CoreV1().PersistentVolumeClaims(storage.Namespace).Update(ctx, storage.updatePvc(data), v1.UpdateOptions{})
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to Update apps Storage: %v", err)
	}
	return nil
}

func (o *storageOperation) delete(ctx context.Context, namespace, name string) error {
	err := o.api.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to delete Storage: %v", err)
	}
	return nil
}

func (o *storageOperation) existsByGroup(ctx context.Context, namespace, group string) (bool, *corev1.PersistentVolumeClaimList, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumeClaims(namespace).List(ctx, v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", types.LabelGroup, group)})
	has, err := o.isExist(storage, err, "Failed to get Storage: %v")
	return has, storage, err
}

func (o *storageOperation) exists(ctx context.Context, namespace, name string) (bool, *corev1.PersistentVolumeClaim, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, v1.GetOptions{})
	has, err := o.isExist(storage, err, "Failed to get Storage: %v")
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

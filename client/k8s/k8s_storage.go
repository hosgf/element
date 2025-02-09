package k8s

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PersistentStorage struct {
	Model
	Storage
}

func (s *Storage) toPersistentStorage(model Model) *PersistentStorage {
	storage := &PersistentStorage{
		Model:   model,
		Storage: *s,
	}
	return storage
}

func (s *PersistentStorage) toPvc() *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      s.Storage.Name,
			Namespace: s.Namespace,
			Labels:    s.labels(),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.PersistentVolumeAccessMode(s.ToAccessMode())},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)},
				Limits:   corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)},
			},
		},
	}
}

func (s *PersistentStorage) updatePvc(data *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	data.Spec.Resources.Limits = corev1.ResourceList{corev1.ResourceStorage: resource.MustParse(s.Size)}
	return data
}

type storageOperation struct {
	*options
}

func (o *storageOperation) Get(ctx context.Context, namespace, name string) (*Storage, error) {
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

func (o *storageOperation) BatchApply(ctx context.Context, model Model, storage []Storage) error {
	if storage == nil || len(storage) < 1 {
		return nil
	}
	model.AllowUpdate = true
	for _, s := range storage {
		if err := o.Apply(ctx, s.toPersistentStorage(model)); err != nil {
			return err
		}
	}
	return nil
}

func (o *storageOperation) Delete(ctx context.Context, namespace, name string) error {
	if o.err != nil {
		return o.err
	}
	if has, _, err := o.exists(ctx, namespace, name); has {
		if err != nil {
			return err
		}
		return o.delete(ctx, namespace, name)
	}
	return nil
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

func (o *storageOperation) exists(ctx context.Context, namespace, name string) (bool, *corev1.PersistentVolumeClaim, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, v1.GetOptions{})
	has, err := o.isExist(storage, err, "Failed to get Storage: %v")
	return has, storage, err
}

func (o *storageOperation) toStorage(datas *corev1.PersistentVolumeClaim) *Storage {
	if datas == nil {
		return nil
	}
	//pods := make([]*Pod, 0, len(datas.Items))
	//for _, p := range datas.Items {
	//	pod := &Pod{
	//		Model:       Model{Namespace: namespace, Name: p.Name},
	//		Status:      string(p.Status.Phase),
	//		Containers:  make([]*Container, 0),
	//		RunningNode: p.Spec.NodeName,
	//	}
	//	pod.setLabels(p.Labels)
	//	for _, c := range p.Spec.Containers {
	//		pod.toContainer(c)
	//	}
	//	pods = append(pods, pod)
	//}
	//return pods
	return &Storage{}
}

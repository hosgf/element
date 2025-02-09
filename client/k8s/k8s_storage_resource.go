package k8s

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PersistentStorageResource struct {
	Model
	Storage
}

func (s *PersistentStorageResource) toPV() *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:      s.Storage.Name,
			Namespace: s.Namespace,
			Labels:    s.labels(),
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				//corev1.ResourceStorage: *resource.NewQuantity(1<<30, resource.BinarySI), // 1Gi
				corev1.ResourceStorage: resource.MustParse(s.Size),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode(s.AccessMode),
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              gconv.String(s.Item),
		},
	}
}

type storageResourceOperation struct {
	*options
}

func (o *storageResourceOperation) Get(ctx context.Context, name string) (*Storage, error) {
	if o.err != nil {
		return nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumes().Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return nil, gerror.NewCodef(gcode.CodeNotImplemented, "Failed to get storages: %v", err)
	}
	return o.toStorage(storage), nil
}

func (o *storageResourceOperation) Exists(ctx context.Context, name string) (bool, error) {
	has, _, err := o.exists(ctx, name)
	return has, err
}

func (o *storageResourceOperation) Create(ctx context.Context, storage *PersistentStorageResource) error {
	if o.err != nil {
		return o.err
	}
	if has, _, err := o.exists(ctx, storage.Storage.Name); has {
		if err != nil {
			return err
		}
		return gerror.NewCodef(gcode.CodeNotImplemented, "Storage: %s, ClaimName: %s 已存在!", storage.Storage.Name, storage.Item)
	}
	return o.create(ctx, storage)
}

func (o *storageResourceOperation) Delete(ctx context.Context, name string) error {
	if o.err != nil {
		return o.err
	}
	if has, _, err := o.exists(ctx, name); has {
		if err != nil {
			return err
		}
		return o.delete(ctx, name)
	}
	return nil
}

func (o *storageResourceOperation) create(ctx context.Context, storage *PersistentStorageResource) error {
	opts := v1.CreateOptions{}
	_, err := o.api.CoreV1().PersistentVolumes().Create(ctx, storage.toPV(), opts)
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to create apps Storage: %v", err)
	}
	return nil
}

func (o *storageResourceOperation) delete(ctx context.Context, name string) error {
	err := o.api.CoreV1().PersistentVolumes().Delete(ctx, name, v1.DeleteOptions{})
	if err != nil {
		return gerror.NewCodef(gcode.CodeNotImplemented, "Failed to delete Storage: %v", err)
	}
	return nil
}

func (o *storageResourceOperation) exists(ctx context.Context, name string) (bool, *corev1.PersistentVolume, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumes().Get(ctx, name, v1.GetOptions{})
	has, err := o.isExist(storage, err, "Failed to get Storage: %v")
	return has, storage, err
}

func (o *storageResourceOperation) toStorage(datas *corev1.PersistentVolume) *Storage {
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

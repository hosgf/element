package k8s

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/hosgf/element/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type storageResourceOperation struct {
	*options
}

func (o *storageResourceOperation) Get(ctx context.Context, name string) (*types.Storage, error) {
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

func (o *storageResourceOperation) Apply(ctx context.Context, storage *PersistentStorageResource) error {
	if o.err != nil {
		return o.err
	}
	res := storage.Resource
	if len(res.Item) < 1 {
		return nil
	}
	if has, _, err := o.exists(ctx, storage.Storage.Name); has {
		if err != nil {
			return err
		}
		if storage.AllowUpdate {
			return nil
		}
		return gerror.NewCodef(gcode.CodeNotImplemented, "Storage: %s, ClaimName: %v 已存在!", storage.Storage.Name, storage.Item)
	}
	return o.create(ctx, storage)
}

func (o *storageResourceOperation) BatchApply(ctx context.Context, model Model, storage []types.Storage) error {
	if storage == nil || len(storage) < 1 {
		return nil
	}
	model.AllowUpdate = true
	for _, s := range storage {
		psr := &PersistentStorageResource{
			Model:   model,
			Storage: s,
		}
		if err := o.Apply(ctx, psr); err != nil {
			return err
		}
	}
	return nil
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

func (o *storageResourceOperation) DeleteByGroup(ctx context.Context, groups ...string) error {
	if o.err != nil {
		return o.err
	}
	var err error
	for _, g := range groups {
		if has, list, err := o.existsByGroup(ctx, g); has {
			if err != nil {
				return err
			}
			if list != nil && list.Size() > 0 {
				for _, i := range list.Items {
					err = o.delete(ctx, i.Name)
				}
			}
		}
	}
	return err
}

func (o *storageResourceOperation) create(ctx context.Context, storage *PersistentStorageResource) error {
	_, err := o.api.CoreV1().PersistentVolumes().Create(ctx, storage.toPv(), v1.CreateOptions{})
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

func (o *storageResourceOperation) existsByGroup(ctx context.Context, group string) (bool, *corev1.PersistentVolumeList, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumes().List(ctx, v1.ListOptions{LabelSelector: fmt.Sprintf("%s=%s", types.LabelGroup, group)})
	has, err := o.isExist(storage, err, "Failed to get Storage: %v")
	return has, storage, err
}

func (o *storageResourceOperation) exists(ctx context.Context, name string) (bool, *corev1.PersistentVolume, error) {
	if o.err != nil {
		return false, nil, o.err
	}
	storage, err := o.api.CoreV1().PersistentVolumes().Get(ctx, name, v1.GetOptions{})
	has, err := o.isExist(storage, err, "Failed to get Storage: %v")
	return has, storage, err
}

func (o *storageResourceOperation) toStorage(datas *corev1.PersistentVolume) *types.Storage {
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
	return &types.Storage{}
}

type Resource struct {
	Name       string   `json:"name,omitempty"`
	Type       string   `json:"type,omitempty"`
	Namespace  string   `json:"namespace,omitempty"`
	Secret     string   `json:"secret,omitempty"`
	SecretName string   `json:"secretName,omitempty"`
	Nodes      []string `json:"nodes,omitempty"`
}

type PersistentStorageResource struct {
	Model
	types.Storage
}

func (s *PersistentStorageResource) toStorageResourceType() StorageResourceType {
	return ToStorageResourceType(s.Resource.Type)
}

func (s *PersistentStorageResource) toPv() *corev1.PersistentVolume {
	spec := corev1.PersistentVolumeSpec{
		Capacity: corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(s.Size),
		},
		AccessModes: []corev1.PersistentVolumeAccessMode{
			corev1.PersistentVolumeAccessMode(s.ToAccessMode()),
		},
		PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
		StorageClassName:              gconv.String(s.Item),
	}
	spec.HostPath = &corev1.HostPathVolumeSource{
		Path: s.GetPath(),
	}
	switch s.ToStorageType() {
	case types.StoragePVC:
		switch s.toStorageResourceType() {
		case StorageResourceRBD:
			volumeMode := corev1.PersistentVolumeBlock
			spec.VolumeMode = &volumeMode
			spec.RBD = s.toRBD()
		}
	}
	return &corev1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:      s.Storage.Name,
			Namespace: s.Namespace,
			Labels:    s.labels(),
		},
		Spec: spec,
	}
}

func (s *PersistentStorageResource) toRBD() *corev1.RBDPersistentVolumeSource {
	r := s.Resource
	if len(r.Item) < 1 {
		return nil
	}
	var rbd corev1.RBDPersistentVolumeSource
	if err := gconv.Struct(r.Item, &rbd); err != nil {
		return nil
	}
	return &rbd
}

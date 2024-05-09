package storageclasses

import (
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewStorageClass is a constructor for a *PersistentVolume object `mountOptions` is an optional parameter and can be nil.
func NewStorageClass(storageClassName, description string, mountOptions []string, reclaimPolicy corev1.PersistentVolumeReclaimPolicy, volumeBindingMode storagev1.VolumeBindingMode) *storagev1.StorageClass {
	annotations := map[string]string{
		"field.cattle.io/description": description,
	}
	// StorageClass object
	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        storageClassName,
			Annotations: annotations,
		},
		MountOptions:      mountOptions,
		ReclaimPolicy:     &reclaimPolicy,
		VolumeBindingMode: &volumeBindingMode,
	}

	return storageClass
}

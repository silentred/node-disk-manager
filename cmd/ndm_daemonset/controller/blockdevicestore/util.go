package blockdevicestore

import (
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

// MergeBlockDeviceData merges the data from BlockDevice resource available in etcd
// with the system generated BlockDevice information
// If the device is in use, then only the capacity, node attributes, path, devlinks
// and state will be updated. This is because, these are the fields relevant even if
// the device is in use.
func MergeBlockDeviceData(newBD, oldBD apis.BlockDevice) *apis.BlockDevice {
	oldBD.TypeMeta = newBD.TypeMeta
	oldBD.ObjectMeta = MergeMetadata(newBD.ObjectMeta, oldBD.ObjectMeta)
	// if the device is in use, only the below fields will be updated.
	if oldBD.Status.ClaimState != apis.BlockDeviceUnclaimed {
		klog.V(4).Infof("device: %s is in use, updating only relevant fields", newBD.Spec.Path)
		oldBD.Spec.NodeAttributes = newBD.Spec.NodeAttributes
		oldBD.Spec.Capacity.Storage = newBD.Spec.Capacity.Storage
		oldBD.Spec.Path = newBD.Spec.Path
		oldBD.Spec.DevLinks = newBD.Spec.DevLinks
		oldBD.Status.State = newBD.Status.State
	} else {
		oldBD.Spec = newBD.Spec
		oldBD.Status = newBD.Status
	}
	return &oldBD
}

// MergeMetadata merges oldMetadata with newMetadata. It takes old metadata and
// update it's value with the help of new metadata.
func MergeMetadata(newMetadata, oldMetadata metav1.ObjectMeta) metav1.ObjectMeta {
	// metadata of older object which contains -
	// - name - no patch required we can use old object.
	// - namespace - no patch required we can use old object.
	// - generateName - no patch required we are not using it.
	// - selfLink - populated by the system we should use old object.
	// - uid - populated by the system we should use old object.
	// - resourceVersion - populated by the system we should use old object.
	// - generation - populated by the system we should use old object.
	// - creationTimestamp - populated by the system we should use old object.
	// - deletionTimestamp - populated by the system we should use old object.
	// - deletionGracePeriodSeconds - populated by the system we should use old object.
	// - labels - we will patch older labels with new labels.
	// - annotations - we will patch older annotations with new annotations.
	// - ownerReferences as ndm-ds is not adding ownerReferences we can go with old object.
	// - initializers ^^^
	// - finalizers ^^^
	// - clusterName - no patch required we can use old object.

	// Patch older label with new label. If there is a new key then it will be added
	// if it is an existing key then value will be overwritten with value from new label
	for key, value := range newMetadata.Labels {
		oldMetadata.Labels[key] = value
	}

	// Patch older annotations with new annotations. If there is a new key then it will be added
	// if it is an existing key then value will be overwritten with value from new annotations
	if oldMetadata.Annotations == nil {
		oldMetadata.Annotations = make(map[string]string)
	}
	for key, value := range newMetadata.Annotations {
		oldMetadata.Annotations[key] = value
	}

	return oldMetadata
}

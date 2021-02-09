package blockdevicestore

import (
	"context"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller/types"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Kubernetes struct {
	NodeAttributes map[string]string
	Namespace      string
	Clientset      client.Client
}

func NewKubernetes(clientset client.Client, ns string, nodeAttrbutes map[string]string) *Kubernetes {
	return &Kubernetes{
		NodeAttributes: nodeAttrbutes,
		Namespace:      ns,
		Clientset:      clientset,
	}
}

func (k *Kubernetes) CreateBlockDevice(blockDevice apis.BlockDevice) error {
	// set namespace on the api resource
	blockDevice.SetNamespace(k.Namespace)

	blockDeviceCopy := blockDevice.DeepCopy()
	err := k.Clientset.Create(context.TODO(), blockDeviceCopy)
	if err == nil {
		klog.Infof("eventcode=%s msg=%s rname=%v",
			"ndm.blockdevice.create.success", "Created blockdevice object in etcd",
			blockDeviceCopy.ObjectMeta.Name)
		return err
	}

	if !errors.IsAlreadyExists(err) {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.blockdevice.create.failure", "Creation of blockdevice object failed",
			err, blockDeviceCopy.ObjectMeta.Name)
		return err
	}

	/*
	 * Creation may fail because resource is already exist in etcd.
	 * This is possible when disk moved from one node to another in
	 * cluster so blockdevice object need to be updated with new Node.
	 */
	err = k.UpdateBlockDevice(blockDevice, nil)
	if err == nil {
		return err
	}

	if !errors.IsConflict(err) {
		klog.Error("Updating of BlockDevice Object failed: ", err)
		return err
	}

	/*
	 * Update might failed due to to resource version mismatch which
	 * can happen if some other entity updating same resource in parallel.
	 */
	err = k.UpdateBlockDevice(blockDevice, nil)
	if err == nil {
		return err
	}
	klog.Error("Update to blockdevice object failed: ", blockDevice.ObjectMeta.Name)
	return nil
}

// UpdateBlockDevice update the BlockDevice resource in etcd
func (k *Kubernetes) UpdateBlockDevice(blockDevice apis.BlockDevice, oldBlockDevice *apis.BlockDevice) error {
	var err error

	blockDeviceCopy := blockDevice.DeepCopy()
	if oldBlockDevice == nil {
		oldBlockDevice = blockDevice.DeepCopy()
		err = k.Clientset.Get(context.TODO(), client.ObjectKey{
			Namespace: oldBlockDevice.Namespace,
			Name:      oldBlockDevice.Name}, oldBlockDevice)
		if err != nil {
			klog.Errorf("eventcode=%s msg=%s : %v, err:%v rname=%v",
				"ndm.blockdevice.update.failure",
				"Failed to update block device : unable to get blockdevice object",
				oldBlockDevice.ObjectMeta.Name, err, blockDeviceCopy.ObjectMeta.Name)
			return err
		}
	}

	blockDeviceCopy = MergeBlockDeviceData(*blockDeviceCopy, *oldBlockDevice)

	err = k.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.blockdevice.update.failure", "Unable to update blockdevice object",
			err, blockDeviceCopy.ObjectMeta.Name)
		return err
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.blockdevice.update.success", "Updated blockdevice object",
		blockDeviceCopy.ObjectMeta.Name)
	return nil
}

// DeactivateBlockDevice API is used to set blockdevice status to "inactive" state in etcd
func (k *Kubernetes) DeactivateBlockDevice(blockDevice apis.BlockDevice) {
	blockDeviceCopy := blockDevice.DeepCopy()
	blockDeviceCopy.Status.State = types.NDMInactive
	err := k.Clientset.Update(context.TODO(), blockDeviceCopy)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v ",
			"ndm.blockdevice.deactivate.failure", "Unable to deactivate blockdevice",
			err, blockDeviceCopy.ObjectMeta.Name)
		return
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.blockdevice.deactivate.success", "Deactivated blockdevice",
		blockDeviceCopy.ObjectMeta.Name)
}

// GetBlockDevice get Disk resource from etcd
func (k *Kubernetes) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	dvr := &apis.BlockDevice{}
	err := k.Clientset.Get(context.TODO(),
		client.ObjectKey{Namespace: k.Namespace, Name: name}, dvr)

	if err != nil {
		klog.Error("Unable to get blockdevice object : ", err)
		return nil, err
	}
	klog.Info("Got blockdevice object : ", name)
	return dvr, nil
}

// DeleteBlockDevice delete the BlockDevice resource from etcd
func (k *Kubernetes) DeleteBlockDevice(name string) {
	blockDevice := &apis.BlockDevice{
		ObjectMeta: metav1.ObjectMeta{
			Labels: make(map[string]string),
			Name:   name,
		},
	}

	err := k.Clientset.Delete(context.TODO(), blockDevice)
	if err != nil {
		klog.Errorf("eventcode=%s msg=%s : %v rname=%v",
			"ndm.blockdevice.delete.failure", "Unable to delete blockdevice object",
			err, name)
		return
	}
	klog.Infof("eventcode=%s msg=%s rname=%v",
		"ndm.blockdevice.delete.success", "Deleted blockdevice object ", name)
}

// ListBlockDeviceResource queries the etcd for the devices
// and returns list of blockdevice resources.
// if listAll = true, all the BlockDevices in the cluster will be listed,
// else only devices present in this node will be listed.
func (k *Kubernetes) ListBlockDeviceResource(listAll bool) (*apis.BlockDeviceList, error) {

	blockDeviceList := &apis.BlockDeviceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "BlockDevice",
			APIVersion: "openebs.io/v1alpha1",
		},
	}
	// create the list options
	var opts []client.ListOption

	// create a new selector
	sel := labels.NewSelector()
	// create a requirement for NDM managed label
	managedRequirement, err := labels.NewRequirement(types.NDMManagedKey, selection.NotEquals, []string{types.FalseString})
	if err != nil {
		return nil, err
	}
	// add the requirements to the selector
	sel.Add(*managedRequirement)

	opts = append(opts, client.MatchingLabelsSelector{Selector: sel})

	if !listAll {
		opts = append(opts, client.MatchingLabels{types.KubernetesHostNameLabel: k.NodeAttributes[types.HostNameKey]})
	}

	err = k.Clientset.List(context.TODO(), blockDeviceList, opts...)
	if err != nil {
		return blockDeviceList, err
	}

	// applying annotation filter, so that blockdevice resources that need not be reconciled are
	// not updated by the daemon
	for i := 0; i < len(blockDeviceList.Items); i++ {
		// if the annotation exists and the value is false, then that blockdevice resource will be removed
		// from the list
		if val, ok := blockDeviceList.Items[i].Annotations[types.OpenEBSReconcile]; ok && util.CheckFalsy(val) {
			blockDeviceList.Items = append(blockDeviceList.Items[:i], blockDeviceList.Items[i+1:]...)
		}
	}
	return blockDeviceList, err
}

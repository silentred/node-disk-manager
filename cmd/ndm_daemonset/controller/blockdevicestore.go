/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/node-disk-manager/pkg/util"
	"k8s.io/klog"
)

// CreateBlockDevice creates the BlockDevice resource in etcd
// This API will be called for each new addDiskEvent
// blockDevice is DeviceResource-CR
func (c *Controller) CreateBlockDevice(blockDevice apis.BlockDevice) error {
	return c.BlockDeviceStore.CreateBlockDevice(blockDevice)
}

// UpdateBlockDevice update the BlockDevice resource in etcd
func (c *Controller) UpdateBlockDevice(blockDevice apis.BlockDevice, oldBlockDevice *apis.BlockDevice) error {
	return c.BlockDeviceStore.UpdateBlockDevice(blockDevice, oldBlockDevice)
}

// DeactivateBlockDevice API is used to set blockdevice status to "inactive" state in etcd
func (c *Controller) DeactivateBlockDevice(blockDevice apis.BlockDevice) {
	c.BlockDeviceStore.DeactivateBlockDevice(blockDevice)
}

// GetBlockDevice get Disk resource from etcd
func (c *Controller) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	return c.BlockDeviceStore.GetBlockDevice(name)
}

// DeleteBlockDevice delete the BlockDevice resource from etcd
func (c *Controller) DeleteBlockDevice(name string) {
	c.BlockDeviceStore.DeleteBlockDevice(name)
}

// ListBlockDeviceResource queries the etcd for the devices
// and returns list of blockdevice resources.
// if listAll = true, all the BlockDevices in the cluster will be listed,
// else only devices present in this node will be listed.
func (c *Controller) ListBlockDeviceResource(listAll bool) (*apis.BlockDeviceList, error) {
	return c.BlockDeviceStore.ListBlockDeviceResource(listAll)
}

// GetExistingBlockDeviceResource returns the existing blockdevice resource if it is
// present in etcd if not it returns nil pointer.
func (c *Controller) GetExistingBlockDeviceResource(blockDeviceList *apis.BlockDeviceList,
	uuid string) *apis.BlockDevice {
	for _, item := range blockDeviceList.Items {
		if uuid == item.ObjectMeta.Name {
			return &item
		}
	}
	return nil
}

// DeactivateStaleBlockDeviceResource deactivates the stale entry from etcd.
// It gets list of resources which are present in system and queries etcd to get
// list of active resources. Active resource which is present in etcd not in
// system that will be marked as inactive.
func (c *Controller) DeactivateStaleBlockDeviceResource(devices []string) {
	listDevices := append(devices, GetActiveSparseBlockDevicesUUID(c.NodeAttributes[HostNameKey])...)
	blockDeviceList, err := c.ListBlockDeviceResource(false)
	if err != nil {
		klog.Error(err)
		return
	}
	for _, item := range blockDeviceList.Items {
		if !util.Contains(listDevices, item.ObjectMeta.Name) {
			c.DeactivateBlockDevice(item)
		}
	}
}

// PushBlockDeviceResource is a utility function which checks old blockdevice resource
// present or not. If it presents in etcd then it updates the resource
// else it creates new blockdevice resource in etcd
func (c *Controller) PushBlockDeviceResource(oldBlockDevice *apis.BlockDevice,
	deviceDetails *DeviceInfo) error {
	deviceDetails.NodeAttributes = c.NodeAttributes
	deviceAPI := deviceDetails.ToDevice()
	if oldBlockDevice != nil {
		return c.UpdateBlockDevice(deviceAPI, oldBlockDevice)
	}
	return c.CreateBlockDevice(deviceAPI)
}

// MarkBlockDeviceStatusToUnknown makes state of all resources owned by node unknown
// This will call as a cleanup process before shutting down.
func (c *Controller) MarkBlockDeviceStatusToUnknown() {
	blockDeviceList, err := c.ListBlockDeviceResource(false)
	if err != nil {
		klog.Error(err)
		return
	}
	for _, item := range blockDeviceList.Items {
		blockDeviceCopy := item.DeepCopy()
		blockDeviceCopy.Status.State = NDMUnknown
		err := c.Clientset.Update(context.TODO(), blockDeviceCopy)
		if err == nil {
			klog.Error("Status marked unknown for blockdevice object: ",
				blockDeviceCopy.ObjectMeta.Name)
		}
	}
}

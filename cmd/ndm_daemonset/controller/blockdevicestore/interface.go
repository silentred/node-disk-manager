package blockdevicestore

import (
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
)

type BlockDeviceStore interface {
	CreateBlockDevice(blockDevice apis.BlockDevice) error
	UpdateBlockDevice(blockDevice apis.BlockDevice, oldBlockDevice *apis.BlockDevice) error
	DeactivateBlockDevice(blockDevice apis.BlockDevice) error
	GetBlockDevice(name string) (*apis.BlockDevice, error)
	DeleteBlockDevice(name string)
	ListBlockDeviceResource(listAll bool) (*apis.BlockDeviceList, error)
	MarkBlockDeviceStatusToUnknown()
}

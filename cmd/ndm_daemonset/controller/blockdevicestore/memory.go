package blockdevicestore

import (
	"encoding/json"
	"sync"

	"github.com/openebs/node-disk-manager/cmd/ndm_daemonset/controller/types"
	apis "github.com/openebs/node-disk-manager/pkg/apis/openebs/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

type Memory struct {
	Namespace string
	// map is name => object
	BlockDevices map[string]apis.BlockDevice
	lock         sync.Mutex
}

func NewMemory(namespace string) *Memory {
	return &Memory{
		Namespace:    namespace,
		BlockDevices: map[string]apis.BlockDevice{},
	}
}

func (m *Memory) CreateBlockDevice(blockDevice apis.BlockDevice) error {
	bs, _ := json.Marshal(blockDevice)
	klog.Infof("CreateBlockDevice %s", string(bs))

	blockDevice.SetNamespace(m.Namespace)
	m.lock.Lock()
	defer m.lock.Unlock()
	m.BlockDevices[blockDevice.Name] = blockDevice
	return nil
}

func (m *Memory) UpdateBlockDevice(blockDevice apis.BlockDevice, oldBlockDevice *apis.BlockDevice) error {
	bs, _ := json.Marshal(blockDevice)
	klog.Infof("UpdateBlockDevice name=%s JSON=%s", oldBlockDevice.Name, string(bs))

	blockDevice.SetNamespace(m.Namespace)
	m.lock.Lock()
	defer m.lock.Unlock()
	m.BlockDevices[oldBlockDevice.Name] = blockDevice
	return nil
}

func (m *Memory) DeactivateBlockDevice(blockDevice apis.BlockDevice) {
	klog.Infof("DeactivateBlockDevice name=%s", blockDevice)
	blockDevice.Status.State = types.NDMInactive
	m.CreateBlockDevice(blockDevice)
}

func (m *Memory) GetBlockDevice(name string) (*apis.BlockDevice, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	b, has := m.BlockDevices[name]
	if has {
		return &b, nil
	}
	errNotFound := errors.NewNotFound(schema.GroupResource{Group: apis.GroupName, Resource: "BlockDevice"}, "BlockDevice")
	return nil, errNotFound
}

func (m *Memory) DeleteBlockDevice(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.BlockDevices, name)
}

func (m *Memory) ListBlockDeviceResource(listAll bool) (*apis.BlockDeviceList, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	list := &apis.BlockDeviceList{}

	for _, item := range m.BlockDevices {
		list.Items = append(list.Items, item)
	}

	return list, nil
}

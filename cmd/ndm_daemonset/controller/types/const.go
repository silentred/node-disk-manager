package types

const (
	// FalseString contains string value of false
	FalseString = "false"
	// TrueString contains string value of true
	TrueString = "true"
	// NDMBlockDeviceKind is the Device kind CR.
	NDMBlockDeviceKind = "BlockDevice"
	// kubernetesLabelPrefix is the prefix for k8s labels
	kubernetesLabelPrefix = "kubernetes.io/"
	// openEBSLabelPrefix is the label prefix for openebs labels
	openEBSLabelPrefix = "openebs.io/"
	// HostNameKey is the key for hostname
	HostNameKey = "hostname"
	// NodeNameKey is the node name label prefix
	NodeNameKey = "nodename"
	// KubernetesHostNameLabel is the hostname label used by k8s
	KubernetesHostNameLabel = kubernetesLabelPrefix + HostNameKey
	// NDMVersion is the CR version.
	NDMVersion = openEBSLabelPrefix + "v1alpha1"
	// reconcileKey is the key used for enable/disable of reconciliation
	reconcileKey = "reconcile"
	// OpenEBSReconcile is used in annotation to check whether CR is to be reconciled or not
	OpenEBSReconcile = openEBSLabelPrefix + reconcileKey
	// NDMNotPartitioned is used to say blockdevice does not have any partition.
	NDMNotPartitioned = "No"
	// NDMPartitioned is used to say blockdevice has some partitions.
	NDMPartitioned = "Yes"
	// NDMActive is constant for active resource status.
	NDMActive = "Active"
	// NDMInactive is constant for inactive resource status.
	NDMInactive = "Inactive"
	// NDMUnknown is constant for resource unknown status.
	NDMUnknown = "Unknown"
	// NDMDeviceTypeKey specifies the block device type
	NDMDeviceTypeKey = "ndm.io/blockdevice-type"
	// NDMManagedKey specifies blockdevice cr should be managed by ndm or not.
	NDMManagedKey = "ndm.io/managed"
)

const (
	// NDMDefaultDiskType will be used to initialize the disk type.
	NDMDefaultDiskType = "disk"
	// NDMDefaultDeviceType will be used to initialize the blockdevice type.
	NDMDefaultDeviceType = "blockdevice"
)

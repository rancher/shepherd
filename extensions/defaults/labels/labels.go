package labels

const (
	EtcdRole         = "node-role.kubernetes.io/etcd"
	ControlplaneRole = "node-role.kubernetes.io/control-plane"
	WorkerRole       = "node-role.kubernetes.io/worker"
	WorkloadSelector = "workload.user.cattle.io/workloadselector"
	RtbOwnerUpdated  = "authz.cluster.cattle.io/rtb-owner-updated"
	InitNode         = "rke.cattle.io/init-node"
	MachineName      = "rke.cattle.io/machine-name"
)

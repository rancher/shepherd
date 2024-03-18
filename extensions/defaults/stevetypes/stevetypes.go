package stevetypes

const (
	Provisioning         = "provisioning.cattle.io.cluster"
	EtcdSnapshot         = "rke.cattle.io.etcdsnapshot"
	FleetCluster         = "fleet.cattle.io.cluster"
	ClusterRoleBinding   = "rbac.authorization.k8s.io.clusterrolebinding"
	PodSecurityAdmission = "management.cattle.io.podsecurityadmissionconfigurationtemplate"
	GlobalRoleBinding    = "management.cattle.io.globalrolebinding"
	Setting              = "management.cattle.io.setting"
	ClusterRepo          = "catalog.cattle.io.clusterrepo"
	Apps                 = "catalog.cattle.io.apps"
	Machine              = "cluster.x-k8s.io.machine"
	Ingress              = "networking.k8s.io.ingress"
	Deployment           = "apps.deployment"
	Daemonset            = "apps.daemonset"
	Service              = "service"
	ServiceAccount       = "serviceaccount"
	Node                 = "node"
	Pod                  = "pod"
	Namespace            = "namespace"
	Configmap            = "configmap"
)

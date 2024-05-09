package annotations

const (
	Machine                = "cluster.x-k8s.io/machine"
	ExternalIp             = "rke.cattle.io/external-ip"
	InternalIp             = "alpha.kubernetes.io/provided-node-ip"
	ControlPlaneLeader     = "control-plane.alpha.kubernetes.io/leader"
	CloudProviderName      = "cloud-provider-name"
	UiSourceRepo           = "catalog.cattle.io/ui-source-repo"
	UiSourceRepoType       = "catalog.cattle.io/ui-source-repo-type"
	ContainerResourceLimit = "field.cattle.io/containerDefaultResourceLimit"
	ProjectId              = "field.cattle.io/projectId"
	Description            = "field.cattle.io/description"
)

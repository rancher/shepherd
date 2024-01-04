package rbac

type Role string

const (
	RestrictedAdmin           Role = "restricted-admin"
	StandardUser              Role = "user"
	ClusterOwner              Role = "cluster-owner"
	ClusterMember             Role = "cluster-member"
	ProjectOwner              Role = "project-owner"
	ProjectMember             Role = "project-member"
	CreateNS                  Role = "create-ns"
	ReadOnly                  Role = "read-only"
	CustomManageProjectMember Role = "projectroletemplatebindings-manage"
	ActiveStatus                   = "active"
	ForbiddenError                 = "403 Forbidden"
)

func (r Role) String() string {
	return string(r)
}

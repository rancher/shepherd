package groupversionresources

import (
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func Node() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
	}
}

func Pod() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
}

func ConfigMap() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}
}

func CustomResourceDefinition() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
}

func Ingress() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}
}

func Project() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "projects",
	}
}

func Role() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    rbacv1.SchemeGroupVersion.Group,
		Version:  rbacv1.SchemeGroupVersion.Version,
		Resource: "roles",
	}
}

func ClusterRole() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    rbacv1.SchemeGroupVersion.Group,
		Version:  rbacv1.SchemeGroupVersion.Version,
		Resource: "clusterroles",
	}
}

func RoleBinding() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    rbacv1.SchemeGroupVersion.Group,
		Version:  rbacv1.SchemeGroupVersion.Version,
		Resource: "rolebindings",
	}
}

func ClusterRoleBinding() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    rbacv1.SchemeGroupVersion.Group,
		Version:  rbacv1.SchemeGroupVersion.Version,
		Resource: "clusterrolebindings",
	}
}

func GlobalRole() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "globalroles",
	}
}

func GlobalRoleBinding() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "globalrolebindings",
	}
}

func ClusterRoleTemplateBinding() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "clusterroletemplatebindings",
	}
}

func RoleTemplate() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "roletemplates",
	}
}

func ProjectRoleTemplateBinding() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "projectroletemplatebindings",
	}
}

func ResourceQuota() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "resourcequotas",
	}
}

func Secret() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}
}

func Service() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}
}

func StorageClass() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "storage.k8s.io",
		Version:  "v1",
		Resource: "storageclasses",
	}
}

func Token() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "management.cattle.io",
		Version:  "v3",
		Resource: "tokens",
	}
}

func PersistentVolumeClaim() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}
}

func PersistentVolume() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumes",
	}
}

func Namespace() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}
}

func Daemonset() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "daemonsets",
	}
}

func Deployment() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
}

func Job() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "batch",
		Version:  "v1",
		Resource: "jobs",
	}
}

func CronJob() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "batch",
		Version:  "v1beta1",
		Resource: "cronjobs",
	}
}

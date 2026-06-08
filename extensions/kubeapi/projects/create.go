package projects

import (
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateProject creates a project using the provided template and wrangler context.
func CreateProject(client *rancher.Client, clusterID string, projectTemplate *v3.Project) (*v3.Project, error) {
	createdProject, err := client.WranglerContext.Mgmt.Project().Create(projectTemplate)
	if err != nil {
		return nil, err
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
			if err != nil {
				return err
			}

			return DeleteProject(adminClient, clusterID, createdProject.Name, true)
		})
	}

	return createdProject, nil
}

// GetProjectByName is a helper function that retrieves a Project from a cluster by name.
func GetProjectByName(client *rancher.Client, clusterID, projectName string) (*v3.Project, error) {
	return client.WranglerContext.Mgmt.Project().Get(clusterID, projectName, metav1.GetOptions{})
}

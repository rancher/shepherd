package projects

import (
	"context"
	"fmt"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateProject is a helper function that uses wrangler context to update an existing project in a cluster
func UpdateProject(client *rancher.Client, updatedProject *v3.Project) (*v3.Project, error) {
	var updated *v3.Project
	var lastErr error
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		current, getErr := GetProjectByName(client, updatedProject.Namespace, updatedProject.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get Project %s: %w", updatedProject.Name, getErr)
			return false, nil
		}

		updatedProject.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.Project().Update(updatedProject)
		if lastErr != nil {
			if k8serrors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}

		return true, nil
	},
	)

	if err != nil {
		return nil, fmt.Errorf("timed out updating Project %s: %w", updatedProject.Name, lastErr)
	}

	return updated, nil
}

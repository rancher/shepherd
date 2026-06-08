package projects

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteProject is a helper function that uses the wrangler context to delete a Project from a cluster.
func DeleteProject(client *rancher.Client, clusterID, projectName string, waitForDeletion bool) error {
	err := client.WranglerContext.Mgmt.Project().Delete(clusterID, projectName, &metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if waitForDeletion {
		return WaitForProjectDeletion(client, clusterID, projectName)
	}

	return nil
}

// WaitForProjectDeletion is a helper function that waits for a Project to be deleted from a cluster.
func WaitForProjectDeletion(client *rancher.Client, clusterID, projectName string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (done bool, err error) {
		_, err = GetProjectByName(client, clusterID, projectName)
		if k8serrors.IsNotFound(err) {
			return true, nil
		}

		if err != nil {
			return false, fmt.Errorf("error checking Project deletion status: %w", err)
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("timed out waiting for Project %s to be deleted: %w", projectName, err)
	}

	return nil
}

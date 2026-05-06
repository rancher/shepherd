package users

import (
	"context"
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// DeleteUser deletes a user by name using wrangler context
func DeleteUser(client *rancher.Client, username string, waitForDelete bool) error {
	err := client.WranglerContext.Mgmt.User().Delete(username, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete user %s: %w", username, err)
	}

	if waitForDelete {
		err = WaitForUserDeletion(client, username)
		if err != nil {
			return fmt.Errorf("timed out waiting for user %s to be deleted: %w", username, err)
		}
	}

	return nil
}

// WaitForUserDeletion polls until a user with the given name is deleted
func WaitForUserDeletion(client *rancher.Client, username string) error {
	return kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, true, func(ctx context.Context) (bool, error) {
		_, err := client.WranglerContext.Mgmt.User().Get(username, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

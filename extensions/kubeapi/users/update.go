package users

import (
	"context"
	"fmt"

	extapi "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// UpdateUser updates an existing user using wrangler context and returns the updated user object
func UpdateUser(client *rancher.Client, user *v3.User) (*v3.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user object is nil")
	}

	var updated *v3.User
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetUserByName(client, user.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get user %s: %w", user.Name, getErr)
			return false, nil
		}
		user.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Mgmt.User().Update(user)
		if lastErr != nil {
			if k8serrors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating user %s: %w", user.Name, lastErr)
	}

	return updated, nil
}

// ChangePasswordForUser updates the password for a given user using wrangler context
func ChangePasswordForUser(client *rancher.Client, userID, currentPassword string, passwordLength int) (string, error) {
	newPassword := generateRandomPassword(passwordLength)
	name := fmt.Sprintf("%s-passwd-change", userID)

	passwordChangeReq := &extapi.PasswordChangeRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: extapi.PasswordChangeRequestSpec{
			CurrentPassword: currentPassword,
			NewPassword:     newPassword,
			UserID:          userID,
		},
	}

	_, err := client.WranglerContext.Ext.PasswordChangeRequest().Create(passwordChangeReq)
	if err != nil {
		return "", fmt.Errorf("failed to change password for user %s: %w", userID, err)
	}

	return newPassword, nil
}

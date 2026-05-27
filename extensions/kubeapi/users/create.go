package users

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	extapi "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// CreateUser creates a user using wrangler context
func CreateUser(client *rancher.Client, user *v3.User) (*v3.User, error) {
	createdUser, err := client.WranglerContext.Mgmt.User().Create(user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user %s: %w", user.Username, err)
	}

	_, err = WaitForUserCreation(client, createdUser.Username)
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for user %s to exist: %w", createdUser.Username, err)
	}

	if client.Session != nil {
		client.Session.RegisterCleanupFunc(func() error {
			return DeleteUser(client, createdUser.Username, true)
		})
	}

	return createdUser, nil
}

// WaitForUserCreation polls until a user with the given username exists and returns the created user
func WaitForUserCreation(client *rancher.Client, username string) (*v3.User, error) {
	var createdUser *v3.User

	err := kwait.PollUntilContextTimeout(context.Background(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		user, err := GetUserByName(client, username)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		createdUser = user
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("timed out waiting for user %s to exist: %w", username, err)
	}

	return createdUser, nil
}

// CreateUserPassword creates an opaque secret for a user password and returns the password.
func CreateUserPassword(client *rancher.Client, username string, passwordLength int) (*corev1.Secret, string, error) {
	generatedPassword := generateRandomPassword(passwordLength)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      username,
			Namespace: UserPasswordSecretNamespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"password": generatedPassword,
		},
	}

	createdSecret, err := client.WranglerContext.Core.Secret().Create(secret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create secret for user %s's password: %w", username, err)
	}

	return createdSecret, generatedPassword, nil
}

func generateRandomPassword(passwordLength int) string {
	password := make([]byte, passwordLength)
	maxInt := big.NewInt(int64(len(PasswordChars)))

	for i := range password {
		n, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			password[i] = 'a'
			continue
		}
		password[i] = PasswordChars[n.Int64()]
	}

	return string(password)
}

// CreateSelfUserRequest retrieves user ID by creating a SelfUser resource using wrangler context
func CreateSelfUserRequest(client *rancher.Client) (string, error) {
	name := namegen.AppendRandomString("selfuser")
	selfUser := &extapi.SelfUser{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	selfUserOutput, err := client.WranglerContext.Ext.SelfUser().Create(selfUser)
	if err != nil {
		return "", fmt.Errorf("failed to create SelfUser %s: %w", name, err)
	}

	userID := selfUserOutput.Status.UserID

	return userID, nil
}

// CreateGroupMembershipRefreshRequest triggers a group membership refresh for a user using wrangler context
func CreateGroupMembershipRefreshRequest(client *rancher.Client, userID string) error {
	name := namegen.AppendRandomString("group-membership-refresh")
	refreshReq := &extapi.GroupMembershipRefreshRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: extapi.GroupMembershipRefreshRequestSpec{
			UserID: userID,
		},
	}

	_, err := client.WranglerContext.Ext.GroupMembershipRefreshRequest().Create(refreshReq)
	if err != nil {
		return fmt.Errorf("failed to create GroupMembershipRefreshRequest %s: %w", name, err)
	}

	return nil
}

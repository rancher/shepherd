package users

import (
	"fmt"

	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	UserPasswordSecretNamespace = "cattle-local-user-passwords"
	PasswordHashAnnotation      = "cattle.io/password-hash"
	PasswordHash                = "pbkdf2sha3512"
	PasswordChars               = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?"
)

// GetUserByName retrieves a user by name using wrangler context
func GetUserByName(client *rancher.Client, username string) (*v3.User, error) {
	user, err := client.WranglerContext.Mgmt.User().Get(username, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}

	return user, nil
}

// ListUsers retrieves all users using wrangler context
func ListUsers(client *rancher.Client) (*v3.UserList, error) {
	users, err := client.WranglerContext.Mgmt.User().List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

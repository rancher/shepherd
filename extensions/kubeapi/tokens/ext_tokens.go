package tokens

import (
	"context"
	"fmt"

	extapi "github.com/rancher/rancher/pkg/apis/ext.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// CreateExtToken creates an ext token using the wrangler context
func CreateExtToken(client *rancher.Client, extToken *extapi.Token) (*extapi.Token, error) {
	createdExtToken, err := client.WranglerContext.Ext.Token().Create(extToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create ext token: %w", err)
	}

	return createdExtToken, nil
}

// GetExtTokenByName retrieves an ext token by name using the wrangler context
func GetExtTokenByName(client *rancher.Client, exttokenName string) (*extapi.Token, error) {
	extToken, err := client.WranglerContext.Ext.Token().Get(exttokenName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ext token %s: %w", exttokenName, err)
	}

	return extToken, nil
}

// ListExtTokens retrieves ext tokens using wrangler context and returns the list of ext tokens
func ListExtTokens(client *rancher.Client, listOpts metav1.ListOptions) (*extapi.TokenList, error) {
	extTokens, err := client.WranglerContext.Ext.Token().List(listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list ext token: %w", err)
	}

	return extTokens, nil
}

// UpdateExtToken updates an existing ext token using wrangler context and returns the updated ext token object
func UpdateExtToken(client *rancher.Client, exttoken *extapi.Token) (*extapi.Token, error) {
	if exttoken == nil {
		return nil, fmt.Errorf("ext token object is nil")
	}

	var updated *extapi.Token
	var lastErr error

	err := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		current, getErr := GetExtTokenByName(client, exttoken.Name)
		if getErr != nil {
			lastErr = fmt.Errorf("failed to get ext token %s: %w", exttoken.Name, getErr)
			return false, nil
		}
		exttoken.ResourceVersion = current.ResourceVersion
		updated, lastErr = client.WranglerContext.Ext.Token().Update(exttoken)
		if lastErr != nil {
			if k8serrors.IsConflict(lastErr) {
				return false, nil
			}
			return false, lastErr
		}
		return true, nil
	})

	if err != nil {
		return nil, fmt.Errorf("timed out updating ext token %s: %w", exttoken.Name, lastErr)
	}

	return updated, nil
}

// DeleteExtToken deletes a ext token by name using wrangler context and waits for the deletion to complete
func DeleteExtToken(client *rancher.Client, exttokenName string, waitForDelete bool) error {
	err := client.WranglerContext.Ext.Token().Delete(exttokenName, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete ext token: %s: %w", exttokenName, err)
	}

	if waitForDelete {
		err = WaitForExtTokenDeletion(client, exttokenName)
		if err != nil {
			return fmt.Errorf("timed out waiting for ext token %s to be deleted: %w", exttokenName, err)
		}
	}
	return nil
}

// WaitForExtTokenDeletion polls until an ext token with the given name is deleted or the timeout is reached
func WaitForExtTokenDeletion(client *rancher.Client, exttokenName string) error {
	return kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveSecondTimeout, defaults.OneMinuteTimeout, false, func(ctx context.Context) (bool, error) {
		_, err := GetExtTokenByName(client, exttokenName)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}

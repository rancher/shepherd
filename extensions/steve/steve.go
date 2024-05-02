package steve

import (
	"context"
	"strings"
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	notFound = "not found"
)

// WaitForSteveResourceDeletion accepts a client, steve resource type, and resource ID, then waits for a steve resource to be deleted
func WaitForSteveResourceDeletion(client *rancher.Client, interval, timeout time.Duration, steveResourceType, steveID string) error {
	err := kwait.PollUntilContextTimeout(context.TODO(), interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		_, err = client.Steve.SteveType(steveResourceType).ByID(steveID)
		if err != nil {
			if strings.Contains(err.Error(), notFound) {
				logrus.Info("Resource was successfully removed!")
				return true, nil
			} else {
				return false, err
			}
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil
}

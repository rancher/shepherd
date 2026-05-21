package tokenregistration

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// GetRegistrationToken polls Rancher for a ClusterRegistrationToken belonging to clusterID
// and returns the first one whose Token and ManifestURL are both populated, or an error on timeout.
func GetRegistrationToken(client *rancher.Client, clusterID string) (*management.ClusterRegistrationToken, error) {
	var populatedToken *management.ClusterRegistrationToken
	start := time.Now()
	logrus.Infof("[%s] retrieving cluster registration token", clusterID)
	lastLog := start

	err := kwait.PollUntilContextTimeout(context.Background(), 2*time.Second, defaults.FifteenMinuteTimeout, true, func(_ context.Context) (done bool, err error) {
		collection, err := client.Management.ClusterRegistrationToken.ListAll(&types.ListOpts{
			Filters: map[string]interface{}{
				"clusterId": clusterID,
			},
		})
		if err != nil {
			logrus.Errorf("[%s] error while listing cluster registration tokens: %v", clusterID, err)
			return false, nil
		}

		if len(collection.Data) == 0 {
			if time.Since(lastLog) >= defaults.OneMinuteTimeout {
				logrus.Warnf("[%s] no cluster registration tokens listed after %s; still waiting", clusterID, time.Since(start).Round(time.Second))
				lastLog = time.Now()
			} else {
				logrus.Debugf("[%s] no cluster registration tokens listed yet", clusterID)
			}
			return false, nil
		}

		for i := range collection.Data {
			t := &collection.Data[i]
			if t.Token != "" && t.ManifestURL != "" {
				populatedToken = t
				return true, nil
			}
		}

		if time.Since(lastLog) >= defaults.OneMinuteTimeout {
			logrus.Warnf("[%s] %d cluster registration token(s) listed but none populated after %s; still waiting", clusterID, len(collection.Data), time.Since(start).Round(time.Second))
			lastLog = time.Now()
		} else {
			logrus.Debugf("[%s] %d cluster registration token(s) listed but Token/ManifestURL not yet populated", clusterID, len(collection.Data))
		}
		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while retrieving registration token for cluster %s after %s: %w", clusterID, defaults.FifteenMinuteTimeout, err)
	}

	logrus.Infof("[%s] cluster registration token populated after %s", clusterID, time.Since(start).Round(time.Second))
	return populatedToken, nil
}

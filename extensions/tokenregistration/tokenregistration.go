package tokenregistration

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/norman/types"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/sirupsen/logrus"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	// registrationTokenPollInterval is the interval between polls for a populated
	// ClusterRegistrationToken.
	registrationTokenPollInterval = 2 * time.Second

	// registrationTokenPollTimeout is the per-cluster ceiling for waiting on
	// Rancher to populate ClusterRegistrationToken.Token and .ManifestURL.
	//
	// At ~500 downstream clusters we measured token-creation lag of
	// p50=9s, p95=78s, p99=141s for the default token and p50=113s, p95=249s,
	// max=315s for the system token. The leader-side cluster-watch handler is
	// chained behind ~15 other Mgmt.Cluster() handlers, so lag scales with N
	// — at 999 clusters the tail extends well past the 500-cluster numbers,
	// so we give 15 minutes of headroom. Tokens that haven't appeared by then
	// almost always indicate a stuck cluster, not lag.
	registrationTokenPollTimeout = 15 * time.Minute

	// quietLogThreshold suppresses the "no tokens listed yet" log line until
	// the poll has been running long enough that absence is actually notable.
	quietLogThreshold = 60 * time.Second
)

// GetRegistrationToken polls Rancher for a ClusterRegistrationToken belonging
// to clusterID and returns the first one whose Token and ManifestURL are both
// populated. An empty list (or one whose entries have not yet been filled in
// by Rancher's controllers) is expected during the first ~10-170s of a
// cluster's life and is not treated as fatal until the overall timeout.
func GetRegistrationToken(client *rancher.Client, clusterID string) (*management.ClusterRegistrationToken, error) {
	var populatedToken *management.ClusterRegistrationToken
	start := time.Now()

	err := kwait.PollUntilContextTimeout(context.Background(), registrationTokenPollInterval, registrationTokenPollTimeout, true, func(_ context.Context) (done bool, err error) {
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
			if time.Since(start) >= quietLogThreshold {
				logrus.Warnf("[%s] no cluster registration tokens listed after %s; still waiting", clusterID, time.Since(start).Round(time.Second))
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

		if time.Since(start) >= quietLogThreshold {
			logrus.Warnf("[%s] %d cluster registration token(s) listed but none populated after %s; still waiting", clusterID, len(collection.Data), time.Since(start).Round(time.Second))
		} else {
			logrus.Debugf("[%s] %d cluster registration token(s) listed but Token/ManifestURL not yet populated", clusterID, len(collection.Data))
		}
		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while retrieving registration token for cluster %s after %s: %w", clusterID, registrationTokenPollTimeout, err)
	}

	logrus.Infof("[%s] cluster registration token populated after %s", clusterID, time.Since(start).Round(time.Second))
	return populatedToken, nil
}

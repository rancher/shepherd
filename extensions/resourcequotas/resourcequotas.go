package resourcequotas

import (
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/states"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// CheckResourceActiveState is a function that uses the Steve API to check if the resource quota is in an active state
func CheckResourceActiveState(client *rancher.Client, resourceQuotaID string) error {
	return kwait.Poll(500*time.Millisecond, 2*time.Minute, func() (done bool, err error) {
		steveResourceQuota, err := client.Steve.SteveType(stevetypes.ResourceQuota).ByID(resourceQuotaID)
		if err != nil {
			return false, err
		} else if steveResourceQuota.State.Name == states.Active {
			return true, nil
		}

		return false, nil
	})
}

package fleet

import (
	"time"

	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	"github.com/rancher/shepherd/clients/rancher"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

const (
	FleetGitRepoResourceType = "fleet.cattle.io.gitrepo"
)

// CreateFleetGitRepo is a "helper" functions that takes a rancher client, and the gitRepo object as parameters. This function
// registers a delete gitRepo fuction to ensure the gitRepo is removed cleanly.
func CreateFleetGitRepo(client *rancher.Client, gitRepo *v1alpha1.GitRepo) (*v1.SteveAPIObject, error) {
	repoObject, err := client.Steve.SteveType(FleetGitRepoResourceType).Create(gitRepo)
	if err != nil {
		return nil, err
	}

	backoff := kwait.Backoff{
		Duration: 1 * time.Second,
		Factor:   1.1,
		Jitter:   0.1,
		Steps:    20,
	}

	err = kwait.ExponentialBackoff(backoff, func() (finished bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			return false, err
		}

		_, err = client.Steve.SteveType(FleetGitRepoResourceType).ByID(repoObject.ID)
		if err != nil {
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	client.Session.RegisterCleanupFunc(func() error {
		client, err = client.ReLogin()
		if err != nil {
			return err
		}

		err = client.Steve.SteveType(FleetGitRepoResourceType).Delete(repoObject)
		if err != nil {
			return err
		}

		return kwait.ExponentialBackoff(backoff, func() (finished bool, err error) {
			client, err = client.ReLogin()
			if err != nil {
				return false, err
			}

			repoObject, _ = client.Steve.SteveType(FleetGitRepoResourceType).ByID(repoObject.ID)
			if repoObject != nil {
				return false, nil
			}

			return true, nil
		})
	})

	return repoObject, nil
}

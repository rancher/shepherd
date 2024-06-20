package vai

import (
	"fmt"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/features"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"k8s.io/apimachinery/pkg/types"
)

const (
	cattleSystemNamespace    = "cattle-system"
	cattleAgentDeployment    = "cattle-cluster-agent"
	clusterRegisterContainer = "cluster-register"
	uiSQLCacheResource       = "ui-sql-cache"
	cattleFeaturesEnvVar     = "CATTLE_FEATURES"
)

func updateSQLCache(adminClient *rancher.Client, value bool) error {
	managementClient := adminClient.Steve.SteveType("management.cattle.io.feature")

	steveCacheFlagResp, err := managementClient.ByID(uiSQLCacheResource)
	if err != nil {
		return err
	}

	_, err = features.UpdateFeatureFlag(adminClient.Steve, steveCacheFlagResp, value)
	if err != nil {
		return err
	}

	errors := pods.StatusPods(adminClient, "local")
	if len(errors) > 1 {
		return fmt.Errorf("error when restarting pods")
	}

	return nil
}

// EnableVaiCaching is the extension that sets all the appropriate global/performance settings, and feature flags
// to enable the vai sql caching.
func EnableVaiCaching(adminClient *rancher.Client) error {
	err := updateSQLCache(adminClient, true)
	if err != nil {
		return err
	}

	adminClient.Session.RegisterCleanupFunc(func() error {
		err := DisableVaiCaching(adminClient)
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

// DisableVaiCaching is the extension that sets all the appropriate global/performance settings, and feature flags
// to disable the vai sql caching.
func DisableVaiCaching(adminClient *rancher.Client) error {
	err := updateSQLCache(adminClient, false)
	if err != nil {
		return err
	}

	return nil
}

func setDownstreamClusterSQLCaching(adminClient *rancher.Client, clusterID, value string) error {
	downStreamController, err := adminClient.WranglerContext.DownStreamClusterWranglerContext(clusterID)
	if err != nil {
		return err
	}

	patchedResource := fmt.Sprintf(`{"spec":{"template":{"spec":{"containers":[{"name":"cluster-register","env":[{"name":"CATTLE_FEATURES","value":"embedded-cluster-api=false,fleet=false,multi-cluster-management=false,multi-cluster-management-agent=true,provisioningv2=false,rke2=false%v"}]}]}}}}`, value)

	_, err = downStreamController.Apps.Deployment().Patch(cattleSystemNamespace, cattleAgentDeployment, types.StrategicMergePatchType, []byte(patchedResource))
	if err != nil {
		return err
	}

	return nil
}

// EnableDownstreamClusterSQLCaching is a function that propagates the caching feature flag to the downstream cluster
func EnableDownstreamClusterSQLCaching(adminClient *rancher.Client, clusterID string) error {
	err := setDownstreamClusterSQLCaching(adminClient, clusterID, fmt.Sprintf(",%v=true", uiSQLCacheResource))
	if err != nil {
		return err
	}

	adminClient.Session.RegisterCleanupFunc(func() error {
		err := DisableDownstreamClusterSQLCaching(adminClient, clusterID)
		if err != nil {
			return err
		}
		return nil
	})

	return nil
}

// DisableDownstreamClusterSQLCaching is a function that removes, if it's there, the ui-sql-cache=true flag.
func DisableDownstreamClusterSQLCaching(adminClient *rancher.Client, clusterID string) error {
	return setDownstreamClusterSQLCaching(adminClient, clusterID, "")
}

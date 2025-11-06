package clusters

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/norman/types"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	apisV1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	v1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/extensions/defaults/stevestates"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	"github.com/rancher/shepherd/pkg/api/scheme"
	"github.com/rancher/shepherd/pkg/wait"
	"github.com/rancher/wrangler/pkg/summary"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kwait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	FleetSteveResourceType               = "fleet.cattle.io.cluster"
	PodSecurityAdmissionSteveResoureType = "management.cattle.io.podsecurityadmissionconfigurationtemplate"
	ProvisioningSteveResourceType        = "provisioning.cattle.io.cluster"
	ErrMsgListDownstreamClusters         = "Couldn't list downstream clusters"

	active                   = "active"
	localcluster             = "fleet-local/local"
	clusterStateUpgrading    = "upgrading" // For imported RKE2 and K3s clusters
	clusterStateUpdating     = "updating"  // For all clusters except imported K3s and RKE2
	clusterErrorStateMessage = "cluster is in error state"
)

// GetV1ProvisioningClusterByName is a helper function that returns the cluster ID by name
func GetV1ProvisioningClusterByName(client *rancher.Client, clusterName string) (string, error) {
	clusterList, err := client.Steve.SteveType(ProvisioningSteveResourceType).List(nil)
	if err != nil {
		return "", err
	}

	for _, cluster := range clusterList.Data {
		if cluster.Name == clusterName {
			return cluster.ID, nil
		}
	}

	return "", nil
}

// GetClusterIDByName is a helper function that returns the cluster ID by name
func GetClusterIDByName(client *rancher.Client, clusterName string) (string, error) {
	clusterList, err := client.Management.Cluster.List(&types.ListOpts{})
	if err != nil {
		return "", err
	}

	for _, cluster := range clusterList.Data {
		if cluster.Name == clusterName {
			return cluster.ID, nil
		}
	}

	return "", nil
}

// GetClusterNameByID is a helper function that returns the cluster ID by name
func GetClusterNameByID(client *rancher.Client, clusterID string) (string, error) {
	clusterList, err := client.Management.Cluster.List(&types.ListOpts{})
	if err != nil {
		return "", err
	}

	for _, cluster := range clusterList.Data {
		if cluster.ID == clusterID {
			return cluster.Name, nil
		}
	}

	return "", nil
}

// IsProvisioningClusterReady is basic check function that would be used for the wait.WatchWait func in pkg/wait.
// This functions just waits until a cluster becomes ready.
func IsProvisioningClusterReady(event watch.Event) (ready bool, err error) {
	cluster := event.Object.(*apisV1.Cluster)
	var updated bool
	ready = cluster.Status.Ready
	for _, condition := range cluster.Status.Conditions {
		if condition.Type == "Updated" && condition.Status == corev1.ConditionTrue {
			updated = true
		}
	}

	return ready && updated, nil
}

// IsHostedProvisioningClusterReady is basic check function that would be used for the wait.WatchWait func in pkg/wait.
// This functions just waits until a hosted cluster becomes ready.
func IsHostedProvisioningClusterReady(event watch.Event) (ready bool, err error) {
	clusterUnstructured := event.Object.(*unstructured.Unstructured)
	cluster := &v3.Cluster{}
	err = scheme.Scheme.Convert(clusterUnstructured, cluster, clusterUnstructured.GroupVersionKind())
	if err != nil {
		return false, err
	}
	for _, cond := range cluster.Status.Conditions {
		if cond.Type == "Ready" && cond.Status == "True" {
			logrus.Infof("Cluster status is active!")
			return true, nil
		}
	}

	return false, nil
}

// CreateRKE1Cluster is a "helper" functions that takes a rancher client, and the rke1 cluster config as parameters. This function
// registers a delete cluster fuction with a wait.WatchWait to ensure the cluster is removed cleanly.
func CreateRKE1Cluster(client *rancher.Client, rke1Cluster *management.Cluster) (*management.Cluster, error) {
	cluster, err := client.Management.Cluster.Create(rke1Cluster)
	if err != nil {
		return nil, err
	}

	err = kwait.Poll(500*time.Millisecond, 2*time.Minute, func() (done bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			return false, err
		}

		_, err = client.Management.Cluster.ByID(cluster.ID)
		if err != nil {
			return false, nil
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}

	client.Session.RegisterCleanupFunc(func() error {
		adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
		if err != nil {
			return err
		}

		clusterResp, err := client.Management.Cluster.ByID(cluster.ID)
		if err != nil {
			return err
		}

		client, err = client.ReLogin()
		if err != nil {
			return err
		}

		err = client.Management.Cluster.Delete(clusterResp)
		if err != nil {
			return err
		}

		watchInterface, err := adminClient.GetManagementWatchInterface(management.ClusterType, metav1.ListOptions{
			FieldSelector:  "metadata.name=" + clusterResp.ID,
			TimeoutSeconds: &defaults.WatchTimeoutSeconds,
		})
		if err != nil {
			return err
		}

		return wait.WatchWait(watchInterface, func(event watch.Event) (ready bool, err error) {
			if event.Type == watch.Error {
				return false, fmt.Errorf("there was an error deleting cluster")
			} else if event.Type == watch.Deleted {
				return true, nil
			}
			return false, nil
		})
	})

	return cluster, nil
}

// CreateK3SRKE2Cluster is a "helper" functions that takes a rancher client, and the rke2 cluster config as parameters. This function
// registers a delete cluster fuction with a wait.WatchWait to ensure the cluster is removed cleanly.
func CreateK3SRKE2Cluster(client *rancher.Client, rke2Cluster *apisV1.Cluster) (*v1.SteveAPIObject, error) {
	cluster, err := client.Steve.SteveType(stevetypes.Provisioning).Create(rke2Cluster)
	if err != nil {
		return nil, err
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (done bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			logrus.Warning("Failed to create client, retrying")
			return false, nil
		}

		_, err = client.Steve.SteveType(stevetypes.Provisioning).ByID(cluster.ID)
		if err != nil {
			return false, nil
		}

		return true, nil
	})

	if err != nil {
		return nil, err
	}

	client.Session.RegisterCleanupFunc(func() error {
		adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
		if err != nil {
			return err
		}

		provKubeClient, err := adminClient.GetKubeAPIProvisioningClient()
		if err != nil {
			return err
		}

		watchInterface, err := provKubeClient.Clusters(cluster.ObjectMeta.Namespace).Watch(context.TODO(), metav1.ListOptions{
			FieldSelector:  "metadata.name=" + cluster.ObjectMeta.Name,
			TimeoutSeconds: &defaults.WatchTimeoutSeconds,
		})

		if err != nil {
			return err
		}

		client, err = client.ReLogin()
		if err != nil {
			return err
		}

		err = client.Steve.SteveType(stevetypes.Provisioning).Delete(cluster)
		if err != nil {
			return err
		}

		return wait.WatchWait(watchInterface, func(event watch.Event) (ready bool, err error) {
			cluster := event.Object.(*apisV1.Cluster)
			if event.Type == watch.Error {
				return false, fmt.Errorf("there was an error deleting cluster")
			} else if event.Type == watch.Deleted {
				return true, nil
			} else if cluster == nil {
				return true, nil
			}
			return false, nil
		})
	})

	return cluster, nil
}

// DeleteKE1Cluster is a "helper" functions that takes a rancher client, and the rke1 cluster ID as parameters to delete
// the cluster.
func DeleteRKE1Cluster(client *rancher.Client, clusterID string) error {
	cluster, err := client.Management.Cluster.ByID(clusterID)
	if err != nil {
		return err
	}

	logrus.Infof("Deleting cluster %s...", cluster.Name)
	err = client.Management.Cluster.Delete(cluster)
	if err != nil {
		return err
	}

	return nil
}

// DeleteK3SRKE2Cluster is a "helper" functions that takes a rancher client, and the non-rke1 cluster ID as parameters to delete
// the cluster.
func DeleteK3SRKE2Cluster(client *rancher.Client, clusterID string) error {
	cluster, err := client.Steve.SteveType(stevetypes.Provisioning).ByID(clusterID)
	if err != nil {
		return err
	}

	err = client.Steve.SteveType(stevetypes.Provisioning).Delete(cluster)
	if err != nil {
		return err
	}

	return nil
}

// UpdateRKE1Cluster is a "helper" functions that takes a rancher client, old rke1 cluster config, and the new rke1 cluster config as parameters.
func UpdateRKE1Cluster(client *rancher.Client, cluster, updatedCluster *management.Cluster) (*management.Cluster, error) {
	logrus.Infof("Updating cluster...")
	newCluster, err := client.Management.Cluster.Update(cluster, updatedCluster)
	if err != nil {
		return nil, err
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, defaults.ThirtyMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			return false, err
		}

		clusterResp, err := client.Management.Cluster.ByID(newCluster.ID)
		if err != nil {
			return false, err
		}

		if clusterResp.State == active {
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

// UpdateK3SRKE2Cluster is a "helper" functions that takes a rancher client, old rke2/k3s cluster config, and the new rke2/k3s cluster config as parameters.
func UpdateK3SRKE2Cluster(client *rancher.Client, cluster *v1.SteveAPIObject, updatedCluster *apisV1.Cluster) (*v1.SteveAPIObject, error) {
	updateCluster, err := client.Steve.SteveType(ProvisioningSteveResourceType).ByID(cluster.ID)
	if err != nil {
		return nil, err
	}

	updatedCluster.ObjectMeta.ResourceVersion = updateCluster.ObjectMeta.ResourceVersion

	logrus.Infof("Updating cluster...")
	cluster, err = client.Steve.SteveType(ProvisioningSteveResourceType).Update(cluster, updatedCluster)
	if err != nil {
		return nil, err
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, defaults.ThirtyMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			return false, err
		}

		clusterResp, err := client.Steve.SteveType(ProvisioningSteveResourceType).ByID(cluster.ID)
		if err != nil {
			return false, err
		}

		clusterStatus := &apisV1.ClusterStatus{}
		err = v1.ConvertToK8sType(clusterResp.Status, clusterStatus)
		if err != nil {
			return false, err
		}

		if clusterResp.ObjectMeta.State.Name == active {
			proxyClient, err := client.Steve.ProxyDownstream(clusterStatus.ClusterName)
			if err != nil {
				return false, err
			}

			_, err = proxyClient.SteveType(pods.PodResourceSteveType).List(nil)
			if err != nil {
				return false, nil
			}

			logrus.Infof("Cluster has been successfully updated!")

			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return nil, err
	}

	return cluster, nil
}

// WaitClusterToBeInUpgrade is a helper function that takes a rancher client, and the cluster id as parameters.
// Waits cluster to be in upgrade state.
// Cluster error states that declare control plane is inaccessible and cluster object modified are ignored.
// Same cluster summary information logging is ignored.
func WaitClusterToBeInUpgrade(client *rancher.Client, clusterID string) (err error) {
	var clusterInfo string
	opts := metav1.ListOptions{
		FieldSelector:  "metadata.name=" + clusterID,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	}

	watchInterface, err := client.GetManagementWatchInterface(management.ClusterType, opts)
	if err != nil {
		return
	}

	checkFuncWaitToBeInUpgrade := func(event watch.Event) (bool, error) {
		acceptableErrorMessages := []string{
			"Cluster health check failed: Failed to communicate with API server during namespace check",
			"the object has been modified",
		}
		clusterUnstructured := event.Object.(*unstructured.Unstructured)
		summarizedCluster := summary.Summarize(clusterUnstructured)

		clusterInfo = logClusterInfoWithChanges(clusterID, clusterInfo, summarizedCluster)

		if summarizedCluster.Transitioning && !summarizedCluster.Error && (summarizedCluster.State == clusterStateUpdating || summarizedCluster.State == clusterStateUpgrading) {
			return true, nil
		} else if summarizedCluster.Error && isClusterInaccessible(summarizedCluster.Message, acceptableErrorMessages) {
			return false, nil
		} else if summarizedCluster.Error && !isClusterInaccessible(summarizedCluster.Message, acceptableErrorMessages) {
			return false, errors.Wrap(err, clusterErrorStateMessage)
		}

		return false, nil
	}
	err = wait.WatchWait(watchInterface, checkFuncWaitToBeInUpgrade)
	if err != nil {
		return
	}

	return
}

// WaitClusterUntilUpgrade is a helper function that takes a rancher client, and the cluster id as parameters.
// Waits until cluster is ready.
// Cluster error states that declare control plane is inaccessible and cluster object modified are ignored.
// Same cluster summary information logging is ignored.
func WaitClusterUntilUpgrade(client *rancher.Client, clusterID string) (err error) {
	var clusterInfo string
	opts := metav1.ListOptions{
		FieldSelector:  "metadata.name=" + clusterID,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	}

	watchInterfaceWaitUpgrade, err := client.GetManagementWatchInterface(management.ClusterType, opts)
	if err != nil {
		return
	}
	checkFuncWaitUpgrade := func(event watch.Event) (bool, error) {
		acceptableErrorMessages := []string{
			"Cluster health check failed: Failed to communicate with API server during namespace check",
			"the object has been modified",
		}
		clusterUnstructured := event.Object.(*unstructured.Unstructured)
		summarizedCluster := summary.Summarize(clusterUnstructured)

		clusterInfo = logClusterInfoWithChanges(clusterID, clusterInfo, summarizedCluster)

		if summarizedCluster.IsReady() {
			return true, nil
		} else if summarizedCluster.Error && isClusterInaccessible(summarizedCluster.Message, acceptableErrorMessages) {
			return false, nil
		} else if summarizedCluster.Error && !isClusterInaccessible(summarizedCluster.Message, acceptableErrorMessages) {
			return false, errors.Wrap(err, clusterErrorStateMessage)

		}

		return false, nil
	}

	err = wait.WatchWait(watchInterfaceWaitUpgrade, checkFuncWaitUpgrade)
	if err != nil {
		return err
	}

	return
}

// WaitForClusterToBeUpgraded is a "helper" functions that takes a rancher client, and the cluster id as parameters. This function
// contains two stages. First stage is to wait to be cluster in upgrade state. And the other is to wait until cluster is ready.
// Cluster error states that declare control plane is inaccessible and cluster object modified are ignored.
// Same cluster summary information logging is ignored.
func WaitClusterToBeUpgraded(client *rancher.Client, clusterID string) (err error) {
	err = WaitClusterToBeInUpgrade(client, clusterID)
	if err != nil {
		return err
	}

	err = WaitClusterUntilUpgrade(client, clusterID)
	if err != nil {
		return err
	}

	return
}

// WaitOnClusterAfterSnapshot waits for a cluster to finish taking a snapshot and return to an active state.
func WaitOnClusterAfterSnapshot(client *rancher.Client, clusterID string) error {
	cluster, err := client.Steve.SteveType(ProvisioningSteveResourceType).ByID(clusterID)
	if err != nil {
		return err
	}

	isTransitioning := cluster.State == nil || cluster.State.Transitioning

	if !isTransitioning {
		err = kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveHundredMillisecondTimeout, defaults.OneMinuteTimeout, true, func(ctx context.Context) (bool, error) {
			cluster, err := client.Steve.SteveType(ProvisioningSteveResourceType).ByID(clusterID)
			if err != nil {
				return false, err
			}

			// note, this intentionally ignores cluster.State.Error, as that can sometimes appear during an upgrade during snapshots.
			if cluster.State == nil {
				return false, nil
			}
			return cluster.State.Transitioning, nil
		})
		if err != nil {
			return err
		}
	}

	err = kwait.PollUntilContextTimeout(context.TODO(), 1*time.Second, defaults.FifteenMinuteTimeout, true, func(ctx context.Context) (bool, error) {
		cluster, err := client.Steve.SteveType(ProvisioningSteveResourceType).ByID(clusterID)
		if err != nil {
			return false, err
		}

		if cluster.State == nil {
			return false, nil
		}
		// note, this intentionally ignores cluster.State.Error, as that can sometimes appear during an upgrade during snapshots.

		return cluster.State.Name == active, nil
	})

	return err
}

func isClusterInaccessible(messages, acceptableErrors []string) (isInaccessible bool) {
	for _, message := range messages {
		if slices.Contains(acceptableErrors, message) {
			isInaccessible = true
			break
		}
	}

	return
}

func logClusterInfoWithChanges(clusterID, clusterInfo string, summary summary.Summary) string {
	newClusterInfo := fmt.Sprintf("ClusterID: %v, Message: %v, Error: %v, State: %v, Transitioning: %v", clusterID, summary.Message, summary.Error, summary.State, summary.Transitioning)

	if clusterInfo != newClusterInfo {
		logrus.Trace(newClusterInfo)
		clusterInfo = newClusterInfo
	}

	return clusterInfo
}

// WatchAndWaitForCluster is function that waits for a cluster to go unactive before checking its active state.
func WatchAndWaitForCluster(client *rancher.Client, steveID string) error {
	var clusterResp *v1.SteveAPIObject
	err := kwait.PollUntilContextTimeout(context.TODO(), 1*time.Second, defaults.TenMinuteTimeout, true, func(ctx context.Context) (done bool, err error) {
		clusterResp, err = client.Steve.SteveType(stevetypes.Provisioning).ByID(steveID)
		if err != nil {
			return false, err
		}
		state := clusterResp.ObjectMeta.State.Name
		return state != stevestates.Active, nil
	})
	if err != nil {
		return err
	}

	adminClient, err := rancher.NewClient(client.RancherConfig.AdminToken, client.Session)
	if err != nil {
		return err
	}
	kubeProvisioningClient, err := adminClient.GetKubeAPIProvisioningClient()
	if err != nil {
		return err
	}

	result, err := kubeProvisioningClient.Clusters(clusterResp.ObjectMeta.Namespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + clusterResp.Name,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(result, IsProvisioningClusterReady)
	return err
}

// GetProvisioningClusterByName is a helper function to get cluster object with the cluster name
func GetProvisioningClusterByName(client *rancher.Client, clusterName string, namespace string) (*apisV1.Cluster, *v1.SteveAPIObject, error) {
	clusterObj, err := client.Steve.SteveType(ProvisioningSteveResourceType).ByID(namespace + "/" + clusterName)
	if err != nil {
		return nil, nil, err
	}

	cluster := new(apisV1.Cluster)
	err = v1.ConvertToK8sType(clusterObj, &cluster)
	if err != nil {
		return nil, nil, err
	}

	return cluster, clusterObj, nil
}

// WaitForActiveCluster is a "helper" function that waits for the cluster to reach the active state.
// The function accepts a Rancher client and a cluster ID as parameters.
func WaitForActiveRKE1Cluster(client *rancher.Client, clusterID string) error {
	err := kwait.Poll(500*time.Millisecond, 30*time.Minute, func() (done bool, err error) {
		client, err = client.ReLogin()
		if err != nil {
			return false, err
		}
		clusterResp, err := client.Management.Cluster.ByID(clusterID)
		if err != nil {
			return false, err
		}
		if clusterResp.State == active {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ListDownstreamClusters is a helper function to get the name of the downstream clusters
func ListDownstreamClusters(client *rancher.Client) (clusterNames []string, err error) {
	clusterList, err := client.Steve.SteveType(ProvisioningSteveResourceType).ListAll(nil)
	if err != nil {
		return nil, errors.Wrap(err, ErrMsgListDownstreamClusters)
	}
	for i, c := range clusterList.Data {
		isLocalCluster := c.ID == localcluster
		if !isLocalCluster {
			clusterNames = append(clusterNames, clusterList.Data[i].Name)
		}
	}
	return
}

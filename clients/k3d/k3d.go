package k3d

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
	apisV3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/clusters"
	"github.com/rancher/shepherd/pkg/config"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/rancher/shepherd/pkg/wait"
	"github.com/rancher/wrangler/pkg/randomtoken"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var importTimeout = int64(60 * 20)

// CreateK3DCluster creates a minimal k3d cluster and returns a rest config for connecting to the newly created cluster.
// If a name is not given a random one will be generated.
func CreateK3DCluster(ts *session.Session, name, hostname string, servers, agents int) (*rest.Config, error) {
	k3dConfig := new(Config)
	config.LoadConfig(ConfigurationFileKey, k3dConfig)

	name = defaultName(name)

	ts.RegisterCleanupFunc(func() error {
		return DeleteK3DCluster(name)
	})

	args := []string{
		"cluster",
		"create",
		name,
		"--no-lb",
		fmt.Sprintf("--servers=%d", servers),
		fmt.Sprintf("--agents=%d", agents),
		"--kubeconfig-update-default=false",
		"--kubeconfig-switch-context=false",
		fmt.Sprintf("--timeout=%d", k3dConfig.createTimeout),
		`--k3s-arg=--kubelet-arg=eviction-hard=imagefs.available<1%,nodefs.available<1%`,
		`--k3s-arg=--kubelet-arg=eviction-minimum-reclaim=imagefs.available=1%,nodefs.available=1%`,
		`--k3s-arg=--disable=traefik`,
		`--k3s-arg=--disable=servicelb`,
		`--k3s-arg=--disable=metrics-serve`,
		`--k3s-arg=--disable=local-storage`,
	}

	if hostname != "" {
		apiHost := fmt.Sprintf("--api-port=%s", hostname)
		args = append(args, apiHost)
	}

	msg, err := exec.Command("k3d", args...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "CreateK3DCluster: "+string(msg))
	}

	configBytes, err := exec.Command("k3d", "kubeconfig", "get", name).Output()
	if err != nil {
		return nil, errors.Wrap(err, "CreateK3DCluster: failed to get kubeconfig for k3d cluster")
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
	if err != nil {
		return nil, errors.Wrap(err, "CreateK3DCluster: failed to parse kubeconfig for k3d cluster")
	}

	return restConfig, nil
}

// DeleteK3DCluster deletes the k3d cluster with the given name. An error is returned if the cluster does not exist.
func DeleteK3DCluster(name string) error {
	return exec.Command("k3d", "cluster", "delete", name).Run()
}

// ImportImage imports an image from docker into the specified k3d cluster. Meant to use local docker images without
// having to setup a registry.
func ImportImage(image, clusterName string) error {
	msg, err := exec.Command("k3d", "image", "import", image, fmt.Sprintf("--cluster=%s", clusterName)).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "ImportImage: "+string(msg))
	}
	return nil
}

// CreateAndImportK3DCluster creates a new k3d cluster and imports it into rancher.
func CreateAndImportK3DCluster(client *rancher.Client, name, image, hostname string, servers, agents int, importImage bool) (*apisV3.Cluster, error) {
	var err error

	name = defaultName(name)

	// create the v3 management cluster
	logrus.Infof("Creating v3 management cluster...")
	v3Cluster := &management.Cluster{
		Name: name,
	}
	v3ClusterResp, err := client.Management.Cluster.Create(v3Cluster)
	if err != nil {
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to create v3 management cluster")
	}

	// create the k3d cluster
	logrus.Infof("Creating K3D cluster...")
	downRest, err := CreateK3DCluster(client.Session, name, hostname, servers, agents)
	if err != nil {
		_ = client.Management.Cluster.Delete(v3ClusterResp)
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to create k3d cluster")
	}

	if importImage {
		logrus.Infof("Importing image to K3D cluster...")
		err = ImportImage(image, name)
		if err != nil {
			return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to import image to k3d cluster")
		}
	}

	// wait for the v3 management cluster to be created
	logrus.Infof("Waiting for v3 management cluster...")
	clusterWatch, err := client.GetManagementWatchInterface(management.ClusterType, metav1.ListOptions{
		FieldSelector:  "metadata.name=" + v3ClusterResp.ID,
		TimeoutSeconds: &importTimeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to watch for v3 management cluster")
	}

	var v3ClusterObj *apisV3.Cluster
	err = wait.WatchWait(clusterWatch, func(event watch.Event) (bool, error) {
		clusterUnstructured := event.Object.(*unstructured.Unstructured)
		cluster := &apisV3.Cluster{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(clusterUnstructured.Object, cluster)
		if err != nil {
			return false, err
		}
		if cluster.Name == v3ClusterResp.ID {
			v3ClusterObj = cluster
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to watch for v3 management cluster")
	}

	// import the k3d cluster
	logrus.Infof("Importing cluster...")
	err = clusters.ImportCluster(client, v3ClusterResp.ID, downRest)
	if err != nil {
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to import cluster")
	}

	// wait for the imported cluster to be ready
	logrus.Infof("Waiting for imported cluster...")
	clusterWatch, err = client.GetManagementWatchInterface(management.ClusterType, metav1.ListOptions{
		FieldSelector:  "metadata.name=" + v3ClusterResp.ID,
		TimeoutSeconds: &importTimeout,
	})
	if err != nil {
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to instantiate the watcher for the cluster")
	}

	checkFunc := clusters.IsImportedClusterReady
	err = wait.WatchWait(clusterWatch, checkFunc)
	if err != nil {
		return nil, errors.Wrap(err, "CreateAndImportK3DCluster: failed to wait for imported cluster ready status")
	}

	return v3ClusterObj, nil
}

// defaultName returns a random string if name is empty, otherwise name is returned unmodified.
func defaultName(name string) string {
	if name == "" {
		name, _ = randomtoken.Generate()
		name = name[:8]
	}

	return name
}

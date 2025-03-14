package harvester

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher"
	steveV1 "github.com/rancher/shepherd/clients/rancher/v1"
	"github.com/rancher/shepherd/extensions/clusters"
)

const (
	HarvesterProviderClusterLabel = "labelSelector=provider.cattle.io=harvester"
)

// KubeConfigOpts is configuration need to get a kubeconfig with SA
type KubeConfigOpts struct {
	CSIClusterRoleName string `yaml:"csiclusterRoleName" json:"csiclusterRoleName" default:"harvesterhci.io:csi-driver"`
	ClusterRoleName    string `yaml:"clusterRoleName" json:"clusterRoleName" default:"harvesterhci.io:cloudprovider"`
	Namespace          string `yaml:"namespace" json:"namespace" default:"default"`
	SaName             string `yaml:"serviceAccountName" json:"serviceAccountName" default:"rancherCluster"`
}

// GetHarvesterSAKubeconfig generates a kubeconfig from the harvester cluster with a custom SA for the
// cluster which will use the kubeconfig. This is typically used when creating a downstream cluster for
// harvester that uses the harvester cloud provider.
func GetHarvesterSAKubeconfig(client *rancher.Client, clusterName string) ([]byte, error) {
	query, err := url.ParseQuery(HarvesterProviderClusterLabel)
	if err != nil {
		return nil, err
	}

	harvesterCluster, err := client.Steve.SteveType(clusters.ProvisioningSteveResourceType).ListAll(query)
	if err != nil {
		return nil, err
	}

	if len(harvesterCluster.Data) == 0 {
		return nil, errors.New("no imported harvester cluster found")
	}

	status := &provv1.ClusterStatus{}
	err = steveV1.ConvertToK8sType(harvesterCluster.Data[0].Status, status)
	if err != nil {
		return nil, err
	}

	opts := KubeConfigOpts{
		CSIClusterRoleName: "harvesterhci.io:csi-driver",
		ClusterRoleName:    "harvesterhci.io:cloudprovider",
		Namespace:          "default",
		SaName:             clusterName,
	}

	bodyContent, err := json.Marshal(opts)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/k8s/clusters/%s/v1/harvester/kubeconfig", client.RancherConfig.Host, status.ClusterName), bytes.NewBuffer(bodyContent))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+client.Management.APIBaseClient.Opts.TokenKey)
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Management.APIBaseClient.Ops.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d error during http request. %s", response.StatusCode, bodyBytes)
	}

	// clean up the string's escapes to be in the expected format
	escapedBody, err := strconv.Unquote(string(bodyBytes))
	return []byte(escapedBody), err
}

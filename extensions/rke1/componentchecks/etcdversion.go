package componentchecks

import (
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/annotations"
	"github.com/rancher/shepherd/extensions/defaults/labels"
	"github.com/rancher/shepherd/extensions/defaults/stevetypes"
	"github.com/rancher/shepherd/pkg/nodes"
	"github.com/sirupsen/logrus"
)

// CheckETCDVersion will check the etcd version on the etcd node in the provisioned RKE1 cluster.
func CheckETCDVersion(client *rancher.Client, nodes []*nodes.Node, clusterID string) ([]string, error) {
	steveClient, err := client.Steve.ProxyDownstream(clusterID)
	if err != nil {
		return nil, err
	}

	nodesList, err := steveClient.SteveType(stevetypes.Node).List(nil)
	if err != nil {
		return nil, err
	}

	var etcdResult []string

	for _, rancherNode := range nodesList.Data {
		externalIP := rancherNode.Annotations[annotations.ExternalIp]
		etcdRole := rancherNode.Labels[labels.EtcdRole] == "true"

		if etcdRole == true {
			for _, node := range nodes {
				if strings.Contains(node.PublicIPAddress, externalIP) {
					command := "docker exec etcd etcdctl version"
					output, err := node.ExecuteCommand(command)
					if err != nil {
						return []string{}, err
					}

					etcdResult = append(etcdResult, output)
					logrus.Infof(output)
				}
			}
		}
	}

	return etcdResult, nil
}

package rancherleader

import (
	"encoding/json"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/configmaps"
)

const (
	KubeSystemNamespace     = "kube-system"
	RancherConfigMap        = "cattle-controllers"
	RancherLeaderAnnotation = "control-plane.alpha.kubernetes.io/leader"
)

// GetRancherLeaderPodName is a helper function to retrieve the name of the rancher leader pod
func GetRancherLeaderPodName(client *rancher.Client) (string, error) {
	configMapList, err := client.Steve.SteveType(configmaps.ConfigMapSteveType).NamespacedSteveClient(KubeSystemNamespace).List(nil)
	if err != nil {
		return "", err
	}

	var leaderAnnotation string
	for _, cm := range configMapList.Data {
		if cm.Name == RancherConfigMap {
			leaderAnnotation = cm.Annotations[RancherLeaderAnnotation]
			break
		}
	}

	var leaderRecord map[string]interface{}
	err = json.Unmarshal([]byte(leaderAnnotation), &leaderRecord)
	if err != nil {
		return "", err
	}

	leaderPodName := leaderRecord["holderIdentity"].(string)

	return leaderPodName, nil
}

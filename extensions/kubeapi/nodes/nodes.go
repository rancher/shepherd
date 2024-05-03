package nodes

import (
	"context"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/defaults/schema/groupversionresources"
	"github.com/rancher/shepherd/pkg/api/scheme"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetNodes returns nodes with metav1.TypeMeta, metav1.ObjectMeta, NodeSpec, and NodeStatus to be used to gather more information from nodes
func GetNodes(client *rancher.Client, clusterID string, listOpts metav1.ListOptions) ([]corev1.Node, error) {
	var nodesList []corev1.Node

	dynamicClient, err := client.GetDownStreamClusterClient(clusterID)
	if err != nil {
		return nil, err
	}

	nodeResource := dynamicClient.Resource(groupversionresources.Node())
	nodes, err := nodeResource.List(context.TODO(), listOpts)
	if err != nil {
		return nil, err
	}

	for _, unstructuredNode := range nodes.Items {
		newNode := &corev1.Node{}
		err := scheme.Scheme.Convert(&unstructuredNode, newNode, unstructuredNode.GroupVersionKind())
		if err != nil {
			return nil, err
		}

		nodesList = append(nodesList, *newNode)
	}

	return nodesList, err
}

// GetNodeIP returns node IP, user needs to pass which type they want ExternalIP, InternalIP, Hostname, check core/v1/types.go
func GetNodeIP(node *corev1.Node, nodeAddressType corev1.NodeAddressType) string {
	nodeAddressList := node.Status.Addresses
	for _, ip := range nodeAddressList {
		if ip.Type == nodeAddressType {
			return ip.Address
		}
	}

	return ""
}

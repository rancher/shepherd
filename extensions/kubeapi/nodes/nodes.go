package nodes

import (
	corev1 "k8s.io/api/core/v1"
)

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

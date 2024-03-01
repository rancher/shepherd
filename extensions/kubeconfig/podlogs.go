package kubeconfig

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8Scheme "k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
)

// GetPodLogs fetches logs from a Kubernetes pod. Buffer size (e.g., '64KB', '8MB', '1GB') influences log reading; an empty string causes no buffering.
func GetPodLogs(client *rancher.Client, clusterID string, podName string, namespace string, bufferSizeStr string) (string, error) {
	var restConfig *restclient.Config

	kubeConfig, err := GetKubeconfig(client, clusterID)
	if err != nil {
		return "", err
	}

	restConfig, err = (*kubeConfig).ClientConfig()
	if err != nil {
		return "", err
	}
	restConfig.ContentConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8Scheme.Scheme)
	restConfig.ContentConfig.GroupVersion = &podGroupVersion
	restConfig.APIPath = apiPath

	restClient, err := restclient.RESTClientFor(restConfig)
	if err != nil {
		return "", err
	}

	req := restClient.Get().Resource("pods").Name(podName).Namespace(namespace).SubResource("log")
	option := &corev1.PodLogOptions{}
	req.VersionedParams(
		option,
		k8Scheme.ParameterCodec,
	)

	stream, err := req.Stream(context.TODO())
	if err != nil {
		return "", fmt.Errorf("error streaming pod logs for pod %s/%s: %v", namespace, podName, err)
	}

	defer stream.Close()

	reader := bufio.NewScanner(stream)

	if bufferSizeStr != "" {
		bufferSize, err := parseBufferSize(bufferSizeStr)
		if err != nil {
			return "", fmt.Errorf("error in parseBufferSize: %v", err)
		}

		buf := make([]byte, bufferSize)
		reader.Buffer(buf, bufferSize)
	}

	var logs string
	for reader.Scan() {
		logs = logs + fmt.Sprintf("%s\n", reader.Text())
	}

	if err := reader.Err(); err != nil {
		return "", fmt.Errorf("error reading pod logs for pod %s/%s: %v", namespace, podName, err)
	}
	return logs, nil
}

// parseBufferSize is a helper function that parses a size string and returns
// the equivalent size in bytes. The provided size string should end with a
// suffix of 'KB', 'MB', or 'GB'. If no suffix is provided, the function will
// return an error.
func parseBufferSize(sizeStr string) (int, error) {
	sizeStr = strings.ToUpper(sizeStr)
	var mult int

	if strings.HasSuffix(sizeStr, "KB") {
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
		mult = 1024
	} else if strings.HasSuffix(sizeStr, "MB") {
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
		mult = 1024 * 1024
	} else if strings.HasSuffix(sizeStr, "GB") {
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
		mult = 1024 * 1024 * 1024
	} else {
		return 0, fmt.Errorf("size must be specified in KB, MB, or GB")
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return 0, err
	}

	return size * mult, nil
}

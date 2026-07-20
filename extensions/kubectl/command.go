package kubectl

import (
	"errors"
	"fmt"
	"strings"

	namegen "github.com/rancher/shepherd/pkg/namegenerator"

	"github.com/rancher/shepherd/clients/rancher"
	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/extensions/workloads"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	corev1 "k8s.io/api/core/v1"
)

const volumeName = "config"

// Command executes the given command on a pod inside the target downstream cluster and returns
// the pod's logs. It runs the command through Rancher's "shell-image" container, which has
// kubectl available and a kubeconfig mounted at /root/.kube, so the command executes against
// the target cluster as a cluster-admin.
//
// The pod is created by submitting a Kubernetes Job (see CreateJobAndRunKubectlCommands), which
// also provisions a throwaway ServiceAccount bound to the cluster-admin role. The function then
// locates the resulting pod by the random suffix appended to the job name and streams its logs.
//
// When yamlContent is non-nil, an init container seeds /config/my-pod.yaml with the provided YAML
// before the main container runs. This is useful for feeding a manifest into the command.
//
// Parameters:
//   - client: the rancher.Client used to talk to the Rancher Management API and proxy the
//     downstream cluster.
//   - yamlContent: optional *management.ImportClusterYamlInput whose YAML is written to the
//     container filesystem via an init container. Pass nil to skip the init container.
//   - clusterID: the ID of the target cluster the command runs against.
//   - command: the command to execute, as an argv-style slice (e.g. []string{"kubectl", "get", "pods"}).
//     Must be non-empty; an empty slice returns an error.
//   - logBufferSize: the size of the buffer used when streaming the pod logs, as a size string
//     (e.g. "64KB", "8MB"). An invalid format is forwarded to the log streamer and may surface
//     as an error.
//
// Returns the pod's log output as a string, or an error if the command is empty, the job or pod
// cannot be created/located, or the logs cannot be streamed. When the underlying job fails, the
// function still attempts to fetch logs because they usually reveal the real cause; the job error
// is wrapped into the returned error only if the logs cannot be retrieved.
func Command(client *rancher.Client, yamlContent *management.ImportClusterYamlInput, clusterID string, command []string, logBufferSize string) (string, error) {

	if len(command) == 0 {
		return "", errors.New("command is empty")
	}

	var user int64
	var group int64
	imageSetting, err := client.Management.Setting.ByID(rancherShellSettingID)
	if err != nil {
		return "", err
	}

	id := namegen.RandStringLower(6)
	jobName := fmt.Sprintf("%v-%v", JobName, id)

	initVolumeMount := []corev1.VolumeMount{
		{
			Name:      volumeName,
			MountPath: "/config",
		},
	}

	volumeMount := []corev1.VolumeMount{
		{
			Name:      volumeName,
			MountPath: "/root/.kube",
		},
	}

	securityContext := &corev1.SecurityContext{
		RunAsUser:  &user,
		RunAsGroup: &group,
	}

	volumes := []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}

	jobTemplate := workloads.NewJobTemplate(jobName, Namespace)

	if yamlContent != nil {
		initContainerCommand := []string{"sh", "-c", fmt.Sprintf("echo \"%s\" > /config/my-pod.yaml", strings.ReplaceAll(yamlContent.YAML, "\"", "\\\""))}
		initContainer := workloads.NewContainer("copy-yaml", imageSetting.Value, corev1.PullAlways, initVolumeMount, nil, initContainerCommand, nil, nil)
		jobTemplate.Spec.Template.Spec.InitContainers = append(jobTemplate.Spec.Template.Spec.InitContainers, initContainer)
	}

	container := workloads.NewContainer(jobName, imageSetting.Value, corev1.PullAlways, volumeMount, nil, command, securityContext, nil)

	jobTemplate.Spec.Template.Spec.Containers = append(jobTemplate.Spec.Template.Spec.Containers, container)
	jobTemplate.Spec.Template.Spec.Volumes = volumes
	// Capture the job error but still try to fetch logs below — they usually reveal
	// the real cause. Only surface jobErr if logs can't be retrieved.
	jobErr := CreateJobAndRunKubectlCommands(clusterID, jobName, jobTemplate, client)

	steveClient, err := client.Steve.ProxyDownstream(clusterID)
	if err != nil {
		return "", err
	}

	pods, err := steveClient.SteveType(pods.PodResourceSteveType).NamespacedSteveClient(Namespace).List(nil)
	if err != nil {
		return "", err
	}

	var podName string
	for _, pod := range pods.Data {
		if strings.Contains(pod.Name, id) {
			podName = pod.Name
			break
		}
	}

	if podName == "" {
		if jobErr != nil {
			return "", fmt.Errorf("kubectl job %s did not produce a pod in namespace %s; job error: %w", jobName, Namespace, jobErr)
		}
		return "", fmt.Errorf("kubectl job %s did not produce a pod in namespace %s", jobName, Namespace)
	}

	podLogs, err := kubeconfig.GetPodLogs(client, clusterID, podName, Namespace, logBufferSize)
	if err != nil {
		if jobErr != nil {
			return "", fmt.Errorf("kubectl job %s failed (job error: %w); failed to stream logs for pod %s/%s: %v", jobName, jobErr, Namespace, podName, err)
		}
		return "", err
	}

	return podLogs, nil
}

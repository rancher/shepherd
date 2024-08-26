package kubectl

import (
	"errors"
	"fmt"
	"net"
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

// Command executes a given command on a Kubernetes pod within a specified cluster using the Rancher Management API and kubectl.
// It optionally sets up an init container to populate a configuration file if yamlContent is provided. The function returns
// the job's logs upon completion. The clusterID identifies the target cluster, and command specifies the command to execute,
// which must not be empty. The logBufferSize defines the size for log output buffering (e.g., "64KB", "8MB", "1GB");
// if empty, the default size is used. An invalid logBufferSize format returns an error. The function returns "StatusOK"
// and the execution logs if successful, or an error detailing the failure.
//
// Parameters:
// - client: Pointer to a rancher.Client used to interact with the Rancher Management API.
// - yamlContent: Optional *management.ImportClusterYamlInput to set up an init container for configuration. If nil, no init container is set up.
// - clusterID: String identifying the target cluster.
// - command: Slice of strings representing the command to execute. Must not be empty.
// - logBufferSize: String representing the log buffer size (e.g., "64KB"). If empty, defaults to the system default size.
//
// Returns:
// - A string containing the logs of the executed job.
// - An error if the command is empty, if there is an issue with setting up the job, or if log retrieval fails.
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
	err = CreateJobAndRunKubectlCommands(clusterID, jobName, jobTemplate, client)
	if err, ok := err.(net.Error); ok && !err.Timeout() {
		return "", err
	}

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
	podLogs, err := kubeconfig.GetPodLogs(client, clusterID, podName, Namespace, logBufferSize)
	if err != nil {
		return "", err
	}

	return podLogs, nil
}

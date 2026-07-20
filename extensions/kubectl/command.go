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

// Command runs the given command on a pod in the target cluster via the Rancher Management API.
// If yamlContent is provided, an init container seeds a config file from it. Returns the job's
// logs on success, or an error (e.g. empty command, bad logBufferSize) otherwise.
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

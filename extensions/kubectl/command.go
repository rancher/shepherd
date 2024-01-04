package kubectl

import (
	"errors"
	"fmt"
	namegen "github.com/rancher/rancher/tests/framework/pkg/namegenerator"
	"net"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/extensions/workloads"
	corev1 "k8s.io/api/core/v1"

	management "github.com/rancher/shepherd/clients/rancher/generated/management/v3"

	"github.com/rancher/shepherd/extensions/workloads/pods"
)

const volumeName = "config"

// Command creates a kubernetes job to execute the provided commands.
// It initializes necessary volume and volume mounts for the job and overrides securityContext to execute the commands.
// If yamlContent is provided, an init container is created to write this content to a file.
// After creating the job it waits for the execution of the job and retrieves the logs of the executed commands
func Command(client *rancher.Client, yamlContent *management.ImportClusterYamlInput, clusterID string, command []string) (string, error) {

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

	steveClient := client.Steve
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
	podLogs, err := kubeconfig.GetPodLogs(client, clusterID, podName, Namespace)
	if err != nil {
		return "", err
	}

	return podLogs, nil
}

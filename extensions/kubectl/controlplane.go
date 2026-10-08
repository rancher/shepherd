package kubectl

import (
	"fmt"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/extensions/kubeconfig"
	"github.com/rancher/shepherd/extensions/workloads"
	"github.com/rancher/shepherd/extensions/workloads/pods"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
	corev1 "k8s.io/api/core/v1"
)

const (
	controlPlaneJobName = "controlplane-command"

	controlPlaneRoleLabel    = "node-role.kubernetes.io/control-plane"
	controlPlaneLegacyLabel  = "node-role.kubernetes.io/controlplane"
	podLogBufferSizeFallback = "64KB"
)

// CommandOnControlPlane runs a command in the host network of a control plane node and returns what it wrote
func CommandOnControlPlane(client *rancher.Client, clusterID string, command []string, logBufferSize string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("the command to run on a control plane node is empty")
	}

	if logBufferSize == "" {
		logBufferSize = podLogBufferSizeFallback
	}

	imageSetting, err := client.Management.Setting.ByID(rancherShellSettingID)
	if err != nil {
		return "", err
	}

	id := namegen.RandStringLower(6)
	jobName := fmt.Sprintf("%v-%v", controlPlaneJobName, id)

	jobTemplate := workloads.NewJobTemplate(jobName, Namespace)
	jobTemplate.Spec.Template.Spec.HostNetwork = true
	jobTemplate.Spec.Template.Spec.Affinity = controlPlaneAffinity()
	jobTemplate.Spec.Template.Spec.Tolerations = []corev1.Toleration{{Operator: corev1.TolerationOpExists}}

	container := workloads.NewContainer(jobName, imageSetting.Value, corev1.PullAlways, nil, nil, command, nil, nil)
	jobTemplate.Spec.Template.Spec.Containers = append(jobTemplate.Spec.Template.Spec.Containers, container)

	jobErr := CreateJobAndRunKubectlCommands(clusterID, jobName, jobTemplate, client)

	steveClient, err := client.Steve.ProxyDownstream(clusterID)
	if err != nil {
		return "", err
	}

	podList, err := steveClient.SteveType(pods.PodResourceSteveType).NamespacedSteveClient(Namespace).List(nil)
	if err != nil {
		return "", err
	}

	var podName string
	for _, pod := range podList.Data {
		if strings.Contains(pod.Name, id) {
			podName = pod.Name
			break
		}
	}

	if podName == "" {
		if jobErr != nil {
			return "", fmt.Errorf("job %s produced no pod in namespace %s; job error: %w", jobName, Namespace, jobErr)
		}

		return "", fmt.Errorf("job %s produced no pod in namespace %s", jobName, Namespace)
	}

	podLogs, err := kubeconfig.GetPodLogs(client, clusterID, podName, Namespace, logBufferSize)
	if err != nil {
		if jobErr != nil {
			return "", fmt.Errorf("job %s failed (job error: %w); streaming logs for pod %s/%s: %v", jobName, jobErr, Namespace, podName, err)
		}

		return "", err
	}

	return podLogs, nil
}

func controlPlaneAffinity() *corev1.Affinity {
	term := func(key string) corev1.NodeSelectorTerm {
		return corev1.NodeSelectorTerm{
			MatchExpressions: []corev1.NodeSelectorRequirement{{
				Key:      key,
				Operator: corev1.NodeSelectorOpExists,
			}},
		}
	}

	return &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					term(controlPlaneRoleLabel),
					term(controlPlaneLegacyLabel),
				},
			},
		},
	}
}

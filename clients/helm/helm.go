package helm

import (
	"os/exec"

	"github.com/pkg/errors"
	"github.com/rancher/shepherd/pkg/session"
)

var helmCmd = ""

func SetHelmCmd(command string) error {
	helmCmd = "helm_v3"
	if command != "" {
		msg, err := exec.Command(command).CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "SetHelmCmd: errored while running `%s` %s", command, string(msg))
		}
		helmCmd = command
	}
	return nil
}

// InstallChart installs a helm chart using helm CLI.
// Send the helm set command strings such as "--set", "installCRDs=true"
// in the args argument to be prepended to the helm install command.
func InstallChart(ts *session.Session, releaseName, helmRepo, namespace, version string, args ...string) error {
	// Register cleanup function
	ts.RegisterCleanupFunc(func() error {
		return UninstallChart(releaseName, namespace)
	})

	// Default helm install command
	commandArgs := []string{
		"install",
		releaseName,
		helmRepo,
		"--namespace",
		namespace,
		"--wait",
	}

	commandArgs = append(commandArgs, args...)

	if version != "" {
		commandArgs = append(commandArgs, "--version", version)
	}

	_, err := execCommand("InstallChart: ", commandArgs)
	return err
}

// UpgradeChart upgrades a helm chart using helm CLI.
// Send the helm set command strings such as "--set", "installCRDs=true"
// in the args argument to be prepended to the helm upgrade command.
func UpgradeChart(ts *session.Session, releaseName, helmRepo, namespace, version string, args ...string) error {
	// Register cleanup function
	ts.RegisterCleanupFunc(func() error {
		return UninstallChart(releaseName, namespace)
	})

	// Default helm upgrade command
	commandArgs := []string{
		"upgrade",
		releaseName,
		helmRepo,
		"--namespace",
		namespace,
		"--wait",
	}

	commandArgs = append(commandArgs, args...)

	if version != "" {
		commandArgs = append(commandArgs, "--version", version)
	}

	_, err := execCommand("UpgradeChart: ", commandArgs)
	return err

}

func GetValues(releaseName, namespace string, args ...string) (string, error) {
	// Default helm upgrade command
	commandArgs := []string{
		"get",
		"values",
		releaseName,
		"--namespace",
		namespace,
	}

	commandArgs = append(commandArgs, args...)

	return execCommand("GetValues: ", commandArgs)
}

// UninstallChart uninstalls a helm chart using helm CLI in a given namespace
// using the releaseName provided.
func UninstallChart(releaseName, namespace string, args ...string) error {
	// Default helm uninstall command
	commandArgs := []string{
		"uninstall",
		releaseName,
		"--namespace",
		namespace,
		"--wait",
	}

	commandArgs = append(commandArgs, args...)

	_, err := execCommand("UninstallChart: ", commandArgs)
	return err
}

// AddHelmRepo adds the specified helm repository using the helm repo add command.
func AddHelmRepo(name, url string) error {
	commandArgs := []string{
		"repo",
		"add",
		name,
		url,
	}

	_, err := execCommand("AddHelmRepo: ", commandArgs)
	return err
}

func execCommand(errMsg string, commandArgs []string) (string, error) {
	msg, err := exec.Command(helmCmd, commandArgs...).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, errMsg+string(msg))
	}

	return string(msg), nil
}

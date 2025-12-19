package helm

import (
	"encoding/json"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/rancher/shepherd/pkg/session"
)

var helmCmd = "helm_v3"

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

	msg, err := exec.Command(helmCmd, commandArgs...).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "InstallChart: "+string(msg))
	}

	return nil
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

	msg, err := exec.Command(helmCmd, commandArgs...).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "UpgradeChart: "+string(msg))
	}

	return nil
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

	msg, err := exec.Command(helmCmd, commandArgs...).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "UninstallChart: "+string(msg))
	}

	return nil
}

// AddHelmRepo adds the specified helm repository using the helm repo add command.
func AddHelmRepo(name, url string) error {
	msg, err := exec.Command(helmCmd, "repo", "add", name, url).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "AddHelmRepo: "+string(msg))
	}

	return nil
}

// UpdateHelmRepo updates the specified helm repository using the helm repo update command.
func UpdateHelmRepo(name string) error {
	msg, err := exec.Command(helmCmd, "repo", "update", name).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "UpdateHelmRepo: "+string(msg))
	}

	return nil
}

// IsReleaseExists checks if a Helm release with the given releaseName exists in the specified namespace.
func IsReleaseExists(releaseName, namespace string) (bool, string, error) {

	commandArgs := []string{
		"list",
		"--namespace",
		namespace,
		"--filter",
		releaseName,
		"--output",
		"json",
	}

	msg, err := exec.Command(helmCmd, commandArgs...).CombinedOutput()
	if err != nil {
		return false, "", errors.Wrap(err, "IsReleaseExists: "+string(msg))
	}

	if string(msg) == "[]" || len(msg) == 0 {
		return false, "", nil
	}

	return true, string(msg), nil
}

// GetAppVersion gets the app version of a given releaseName in the specified namespace
func GetAppVersion(releaseName, namespace string) (string, error) {
	exists, msg, err := IsReleaseExists(releaseName, namespace)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New("release does not exist")
	}

	var releases []struct {
		Name       string `json:"name"`
		AppVersion string `json:"app_version"`
	}

	if err := json.Unmarshal([]byte(msg), &releases); err != nil {
		return "", errors.Wrap(err, "GetAppVersion: failed to unmarshal json")
	}

	for _, r := range releases {
		if r.Name == releaseName {
			return r.AppVersion, nil
		}
	}

	return "", errors.New("release not found in list output")
}

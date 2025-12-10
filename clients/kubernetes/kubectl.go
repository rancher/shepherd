package kubernetes

import (
	"os/exec"

	"github.com/pkg/errors"
)

// ApplyYAML applies a Kubernetes YAML file using kubectl.
func ApplyYAML(yamlPath string) error {
	cmd := exec.Command("kubectl", "apply", "-f", yamlPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "ApplyYAML: "+string(output))
	}
	return nil
}

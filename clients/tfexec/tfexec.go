package tfexec

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
	namegen "github.com/rancher/shepherd/pkg/namegenerator"
)

const (
	debugFlag            = "--trace"
	skipCleanupFlag      = "--skip-cleanup"
	tfVersionConstraint  = "<= 1.5.7"
	tfWorkspacePrefix    = "shepherd-"
	tfPlanFilePathPrefix = "tfexec_plan_"
)

// A representation of shepherd's tfexec client which wraps the upstream tfexec.Terraform struct client
// and its own custom Config struct
type Client struct {
	// Client used to access Terraform
	Terraform       *tfexec.Terraform
	TerraformConfig *Config
}

// NewClient loads the tfexec client's config, initializes a new instance of the upstream Terraform Client and/or returns an error
func NewClient() (*Client, error) {
	tfConfig := TerraformConfig()

	tf, err := tfexec.NewTerraform(tfConfig.WorkingDir, tfConfig.ExecPath)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create new Terraform instance")
	}

	c := &Client{
		Terraform:       tf,
		TerraformConfig: tfConfig,
	}
	return c, nil
}

// InitTerraform initializes the Terraform module (effectively the `terraform init` command)
// This function should be called if the Terraform module in use has not been initialized before
// Returns an error or nil
func (c *Client) InitTerraform(opts ...tfexec.InitOption) error {
	err := c.Terraform.Init(context.Background(), opts...)
	if err != nil {
		return errors.Wrap(err, "InitTerraform: ")
	}
	return nil
}

// InstallTfVersion uses hc-install in order to install the desired Terraform version
// Returns the execPath of the newly installed Terraform binary and/or an error
func InstallTfVersion(tfVersion string) (string, error) {
	v, err := version.NewVersion(tfVersion)
	if err != nil {
		return "", errors.Wrap(err, "InstallTfVersion: ")
	}
	tfConstraint, err := version.NewConstraint(tfVersionConstraint)
	if err != nil {
		return "", errors.Wrap(err, "InstallTfVersion: ")
	}
	if !tfConstraint.Check(v) {
		return "", errors.New("InstallTfVersion: version string '" + tfVersion + "'did not meet constraint criteria of " + tfVersionConstraint)
	}

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(tfVersion)),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "InstallTfVersion: failed to install Terraform")
	}

	return execPath, nil
}

// ShowState effectively runs the `terraform show` command which reads the default state path and outputs the current state
// Returns the state representation as a struct or an error if any
func (c *Client) ShowState(opts ...tfexec.ShowOption) (*tfjson.State, error) {
	state, err := c.Terraform.Show(context.Background(), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "ShowState: ")
	}
	return state, nil
}

// WorkspaceExists lists the existing workspaces for the module in use and verifies that the
// given workspace name exists in the list (effectively runs the `terraform workspace list` command)
// Returns a bool signaling if the workspace exists and an error if any
func (c *Client) WorkspaceExists(workspace string) (bool, error) {
	wsList, _, err := c.Terraform.WorkspaceList(context.Background())
	if err != nil {
		return false, errors.Wrap(err, "WorkspaceExists: ")
	}
	for _, ws := range wsList {
		if ws == workspace {
			return true, nil
		}
	}
	return false, nil
}

// SetupWorkspace will effectively select an existing workspace or create a new one based on
// if the given workspace name already exists. It effectively runs one of the following scenarios:
//   - workspaceName input provided in Config > terraform workspace list > workspaceName in list > terraform workspace select workspaceName
//   - workspaceName input provided in Config > terraform workspace list > workspaceName NOT in list > terraform workspace new workspaceName
//   - workspaceName input NOT provided in Config > terraform workspace new "shepherd-<randomString>"
//
// Returns an error if any
func (c *Client) SetupWorkspace(opts ...tfexec.WorkspaceNewCmdOption) error {
	ws := c.TerraformConfig.WorkspaceName

	var err error
	var wsExists bool
	if ws != "" {
		wsExists, err = c.WorkspaceExists(ws)
		if err != nil {
			return errors.Wrap(err, "SetupWorkspace: ")
		}
		if wsExists {
			err = c.Terraform.WorkspaceSelect(context.Background(), ws)
		} else {
			err = c.Terraform.WorkspaceNew(context.Background(), ws, opts...)
		}
		c.TerraformConfig.WorkspaceName = ws
	} else {
		c.TerraformConfig.WorkspaceName = tfWorkspacePrefix + namegen.RandStringLower(5)
		err = c.Terraform.WorkspaceNew(context.Background(), c.TerraformConfig.WorkspaceName, opts...)
	}
	if err != nil {
		return errors.Wrap(err, "SetupWorkspace: ")
	}
	return nil
}

// PlanJSON generates a JSON-formatted terraform planfile, it effectively runs the `terraform plan -json“ command.
// Takes an io.Writer and any number of tfexec.PlanOptions. The writer is used to determine where to send any output, all PlanOptions will be respected.
// If no -out option is passed, and the PlanFilePath has NOT been configured, then it will generate a timestamped PlanFilePath
// using the PlanOpts.OutDir, tfPlanFilePathPrefix string, and timestamp.
// If no -var-file option is passed, and the VarFilePath has been configured, then it will use the VarFilePath.
// Returns an error if any
func (c *Client) PlanJSON(w io.Writer, opts ...tfexec.PlanOption) error {
	hasOutOpt, hasVarFileOpt := false, false
	var parsedOpts []tfexec.PlanOption
	for _, opt := range opts {
		switch opt.(type) {
		case *tfexec.OutOption:
			hasOutOpt = true
		case *tfexec.VarFileOption:
			hasVarFileOpt = true
		default:
			parsedOpts = append(parsedOpts, tfexec.PlanOption(opt))
		}
	}

	if !hasOutOpt {
		if c.TerraformConfig.PlanFilePath == "" {
			if c.TerraformConfig.PlanOpts == nil || c.TerraformConfig.PlanOpts.OutDir == "" {
				return errors.New("PlanJSON: PlanOpts configuration field is nil or PlanOpts.OutDir is empty. Could not generate PlanFilePath for output")
			}
			c.TerraformConfig.PlanFilePath = c.TerraformConfig.PlanOpts.OutDir + "/" + tfPlanFilePathPrefix + time.Now().Format(time.RFC3339) + ".tfplan"
		}
		parsedOpts = append(parsedOpts, tfexec.PlanOption(tfexec.Out(c.TerraformConfig.PlanFilePath)))
	}
	if !hasVarFileOpt && c.TerraformConfig.VarFilePath != "" {
		parsedOpts = append(parsedOpts, tfexec.PlanOption(tfexec.VarFile(c.TerraformConfig.VarFilePath)))
	}

	_, err := c.Terraform.PlanJSON(context.Background(), w, parsedOpts...)
	if err != nil {
		return errors.Wrap(err, "PlanJSON: ")
	}
	return nil
}

// ApplyPlanJSON effectively runs the `terraform apply -json` command.
// Takes an io.Writer and any number of tfexec.ApplyOptions. The writer is used to determine where to send any output, all ApplyOptions will be respected.
// If a tfexec.DirOrPlanOption is NOT passed, and there is a configured PlanFilePath, then that path will be used.
// Returns an error if any
func (c *Client) ApplyPlanJSON(w io.Writer, opts ...tfexec.ApplyOption) error {
	hasPlanFileOpt := false
	var parsedOpts []tfexec.ApplyOption
	for _, opt := range opts {
		switch opt.(type) {
		case *tfexec.DirOrPlanOption:
			hasPlanFileOpt = true
		default:
			parsedOpts = append(parsedOpts, tfexec.ApplyOption(opt))
		}
	}

	if !hasPlanFileOpt {
		if c.TerraformConfig.PlanFilePath == "" {
			return errors.New("ApplyPlanJSON: No PlanFilePath or tfexec.DirOrPlanOption was provided")
		}
		parsedOpts = append(parsedOpts, tfexec.ApplyOption(tfexec.DirOrPlan(c.TerraformConfig.PlanFilePath)))
	}

	err := c.Terraform.ApplyJSON(context.Background(), w, parsedOpts...)
	if err != nil {
		return errors.Wrap(err, "ApplyPlanJSON: ")
	}

	return nil
}

// Output effectively runs the `terraform output -json` command
// It takes any number of tfexec.OutputOptions as inputs and respects them.
// This function will parse the machine-readable JSON into a []map[string]any that can be further manipulated.
// Returns a []map[string]any of the generated output, and an error if any
func (c *Client) Output(opts ...tfexec.OutputOption) ([]map[string]any, error) {
	outputs, err := c.Terraform.Output(context.Background(), opts...)
	if err != nil {
		return nil, errors.Wrap(err, "Output: ")
	}

	var parsedOutput []map[string]any
	for key, output := range outputs {
		var val any

		err := json.Unmarshal(output.Value, &val)
		if err != nil {
			return nil, errors.Wrap(err, "Output: ")
		}

		tempMap := map[string]any{key: val}
		parsedOutput = append(parsedOutput, tempMap)
	}
	return parsedOutput, nil
}

// WorkingDir simply returns the configured terraform working directory for terraform
// The WorkingDir is the directory that terraform is targetting when it runs a command
func (c *Client) WorkingDir() string {
	return c.Terraform.WorkingDir()
}

// DestroyJSON effectively runs the `terraform destroy -json` command
// Takes an io.Writer and any number of tfexec.DestroyOptions. The writer is used to determine where to send any output, all DestroyOptions will be respected.
// If the -var-file option is not passed, and a VarFilePath is configured, then it will use this path for the -var-file option.
// Returns an error if any
func (c *Client) DestroyJSON(w io.Writer, opts ...tfexec.DestroyOption) error {
	hasVarFileOpt := false
	var parsedOpts []tfexec.DestroyOption
	for _, opt := range opts {
		switch opt.(type) {
		case *tfexec.VarFileOption:
			hasVarFileOpt = true
		default:
			parsedOpts = append(parsedOpts, tfexec.DestroyOption(opt))
		}
	}

	if !hasVarFileOpt {
		if c.TerraformConfig.VarFilePath != "" {
			parsedOpts = append(parsedOpts, tfexec.DestroyOption(tfexec.VarFile(c.TerraformConfig.VarFilePath)))
		}
	}

	err := c.Terraform.DestroyJSON(context.Background(), w, parsedOpts...)
	if err != nil {
		return errors.Wrap(err, "Destroy: ")
	}

	return nil
}

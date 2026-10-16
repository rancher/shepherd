package ranchercli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
)

const (
	JSONOutput         = "json"
	EnvelopeAPIVersion = "cli.rancher.io/v1"
	ExitNotFound       = 4
	ExitRBAC           = 6
	ExitTLS            = 10
	ExitAccessDenied   = 12
	configDirEnv       = "RANCHER_CLI_CONFIG_DIR"
	credentialStoreEnv = "RANCHER_CLI_CREDENTIAL_STORE"
	caFileName         = "ca.pem"
)

// Result is the captured output of a CLI invocation.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// ExitError is returned when the CLI exits with a non-zero status.
type ExitError struct {
	Args     []string
	ExitCode int
	Stdout   string
	Stderr   string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("rancher %s exited with code %d: %s", strings.Join(e.Args, " "), e.ExitCode, strings.TrimSpace(e.Stderr))
}

// Envelope is the JSON document emitted by the CLI when run with -o json.
type Envelope struct {
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Data       json.RawMessage `json:"data"`
}

// ExitCodeOf returns the CLI exit code carried by err, if any.
func ExitCodeOf(err error) (int, bool) {
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode, true
	}

	return 0, false
}

// IsExitCode reports whether err is a CLI failure with the given exit code.
func IsExitCode(err error, code int) bool {
	got, ok := ExitCodeOf(err)
	return ok && got == code
}

// CLIClient drives the new Rancher CLI against an isolated config directory.
type CLIClient struct {
	Name            string
	ConfigDirectory string
	CAFile          string
}

// CLIOptions configures NewCLIClient.
type CLIOptions struct {
	Name    string
	Host    string
	Token   string
	CAFile  string
	CACerts string
	Session *session.Session
}

// NewCLIClient creates an isolated config directory and logs into the Rancher server.
func NewCLIClient(opts CLIOptions) (*CLIClient, error) {
	configDir, err := os.MkdirTemp("", "rancher-cli-")
	if err != nil {
		return nil, fmt.Errorf("create CLI config directory: %w", err)
	}

	c := &CLIClient{Name: opts.Name, ConfigDirectory: configDir, CAFile: opts.CAFile}

	if c.CAFile == "" && opts.CACerts != "" {
		c.CAFile = filepath.Join(configDir, caFileName)
		if err := os.WriteFile(c.CAFile, []byte(opts.CACerts), 0o600); err != nil {
			_ = c.Close()
			return nil, fmt.Errorf("write CA file: %w", err)
		}
	}

	args := []string{"auth", "login", opts.Name, "--url", serverURL(opts.Host), "--token-stdin"}
	if c.CAFile != "" {
		args = append(args, "--cacert", c.CAFile)
	}

	args = append(args, "-o", JSONOutput)

	logrus.Infof("Logging into Rancher server %s with the CLI", opts.Host)

	if _, err := c.execute(context.Background(), strings.NewReader(opts.Token), args...); err != nil {
		_ = c.Close()

		if IsExitCode(err, ExitTLS) {
			return nil, fmt.Errorf("%w (TLS verification failed; provide CAFile or CACerts for the Rancher server)", err)
		}

		return nil, err
	}

	if opts.Session != nil {
		opts.Session.RegisterCleanupFunc(c.Close)
	}

	return c, nil
}

// Run executes the CLI with args and returns an *ExitError on a non-zero exit.
func (c *CLIClient) Run(ctx context.Context, args ...string) (Result, error) {
	return c.execute(ctx, nil, args...)
}

// RunJSON executes the CLI with -o json and decodes and validates the envelope.
func (c *CLIClient) RunJSON(ctx context.Context, args ...string) (*Envelope, error) {
	result, err := c.execute(ctx, nil, append(args, "-o", JSONOutput)...)
	if err != nil {
		return nil, err
	}

	envelope := &Envelope{}
	if err := json.Unmarshal([]byte(result.Stdout), envelope); err != nil {
		return nil, fmt.Errorf("decode CLI envelope for 'rancher %s': %w", strings.Join(args, " "), err)
	}

	if envelope.APIVersion != EnvelopeAPIVersion {
		return nil, fmt.Errorf("unexpected CLI envelope apiVersion %q, want %q", envelope.APIVersion, EnvelopeAPIVersion)
	}

	return envelope, nil
}

// Close removes the config directory, including stored credentials and the generated CA file.
func (c *CLIClient) Close() error {
	if c.ConfigDirectory == "" {
		return nil
	}

	return os.RemoveAll(c.ConfigDirectory)
}

func (c *CLIClient) execute(ctx context.Context, stdin io.Reader, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, "rancher", args...)
	cmd.Env = append(os.Environ(),
		configDirEnv+"="+c.ConfigDirectory,
		credentialStoreEnv+"=file",
	)

	if stdin != nil {
		cmd.Stdin = stdin
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{Stdout: stdout.String(), Stderr: stderr.String()}

	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result, &ExitError{Args: args, ExitCode: result.ExitCode, Stdout: result.Stdout, Stderr: result.Stderr}
	}

	return result, fmt.Errorf("run 'rancher %s': %w", strings.Join(args, " "), err)
}

func serverURL(host string) string {
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}

	return "https://" + host
}

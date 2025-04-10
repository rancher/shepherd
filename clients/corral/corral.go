package corral

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	debugFlag               = "--debug"
	skipCleanupFlag         = "--skip-cleanup"
	corralPrivateSSHKey     = "corral_private_key"
	corralPublicSSHKey      = "corral_public_key"
	corralRegistryIP        = "registry_ip"
	corralRegistryPrivateIP = "registry_private_ip"
)

// GetCorralEnvVar gets corral environment variables
func GetCorralEnvVar(corralName, envVar string) (string, error) {
	msg, err := exec.Command("corral", "vars", corralName, envVar).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "GetCorralEnvVar: "+string(msg))
	}

	corralEnvVar := string(msg)
	corralEnvVar = strings.TrimSuffix(corralEnvVar, "\n")
	return corralEnvVar, nil
}

// SetupCorralConfig sets the corral config vars. It takes a map[string]string as a parameter; the key is the value and the value the value you are setting
// For example we are getting the aws config vars to build a corral from aws.
// results := aws.AWSCorralConfigVars()
// err := corral.SetupCorralConfig(results)
func SetupCorralConfig(configVars map[string]string, configUser string, configSSHPath string) error {
	msg, err := exec.Command("corral", "config", "--user_id", configUser, "--public_key", configSSHPath).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Unable to set configuraion: "+string(msg))
	}
	for variable, value := range configVars {
		msg, err := exec.Command("corral", "config", "vars", "set", variable, value).CombinedOutput()
		if err != nil {
			return errors.Wrap(err, "SetupCorralConfig: "+string(msg))
		}
	}
	return nil
}

// SetCustomRepo sets a custom repo for corral to use. It takes a string as a parameter which is the repo you want to use
func SetCustomRepo(repo string) error {
	msg, err := exec.Command("git", "clone", repo, "corral-packages").CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed to git clone remote repo: "+string(msg))
	}
	makemsg, err := exec.Command("make", "init", "-C", "corral-packages").CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed to git clone remote repo: "+string(makemsg))
	}
	makebuildmsg, err := exec.Command("make", "build", "-C", "corral-packages").CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "Failed to git clone remote repo: "+string(makebuildmsg))
	}
	logrus.Infof("Successfully set custom repo: %s", repo)
	return nil
}

// CreateCorral creates a corral taking the corral name, the package path, and a debug set so if someone wants to view the
// corral create log
func CreateCorral(ts *session.Session, corralName, packageName string, debug, cleanup bool) ([]byte, error) {
	command, err := startCorral(ts, corralName, packageName, debug, cleanup)
	if err != nil {
		return nil, err
	}

	return runAndWaitOnCommand(command)
}

func runAndWaitOnCommand(command *exec.Cmd) ([]byte, error) {
	err := command.Wait()
	var msg []byte
	if command.Stdout != nil {
		msg = command.Stdout.(*bytes.Buffer).Bytes()
	}

	if msg != nil {
		logrus.Infof("Stdout: %s", string(msg))
	}

	return msg, errors.Wrap(err, "Debug: "+string(msg))
}

func startCorral(ts *session.Session, corralName, packageName string, debug, cleanup bool) (*exec.Cmd, error) {
	ts.RegisterCleanupFunc(func() error {
		return DeleteCorral(corralName)
	})

	args := []string{"create"}

	if !cleanup {
		args = append(args, skipCleanupFlag)
	}
	if debug {
		args = append(args, debugFlag)
	}

	args = append(args, corralName, packageName)
	logrus.Infof("Creating corral with the following parameters: %v", args)

	cmdToRun := exec.Command("corral", args...)

	// create a buffer for stdout/stderr so we can read from it later. commands initiate this to nil by default.
	var b bytes.Buffer
	cmdToRun.Stdout = &b
	cmdToRun.Stderr = &b
	err := cmdToRun.Start()
	if err != nil {
		return nil, err
	}

	// this ensures corral is completely initiated. Otherwise, race conditions occur.
	err = waitForCorralConfig(corralName)
	if err != nil {
		return nil, err
	}

	return cmdToRun, err
}

func waitForCorralConfig(corralName string) error {
	backoff := wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   1.1,
		Jitter:   0.1,
		Steps:    10,
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	corralOSPath := homeDir + "/.corral/corrals/" + corralName + "/corral.yaml"

	return wait.ExponentialBackoff(backoff, func() (finished bool, err error) {
		fileStat, err := os.Stat(corralOSPath)
		if err != nil {
			return false, nil
		}

		if fileStat == nil {
			return false, nil
		}

		fileContents, err := os.ReadFile(corralOSPath)
		if err != nil {
			return false, nil
		}

		if fileContents == nil {
			return false, nil
		}

		if len(string(fileContents)) <= 0 {
			return false, nil
		}

		return true, err
	})
}

// CreateMultipleCorrals creates corrals taking the corral name, the package path, and a debug set so if someone wants to view the
// corral create log. Using this function implies calling WaitOnCorralWithCombinedOutput to get the output once finished.
func CreateMultipleCorrals(ts *session.Session, commands []Args, debug, cleanup bool) ([][]byte, error) {
	var waitGroup sync.WaitGroup

	var msgs [][]byte
	var errStrings []string

	for _, currentCommand := range commands {
		// break out of any error that comes up before we run the waitGroup, to avoid running if we're already in an error state.
		for key, value := range currentCommand.Updates {
			logrus.Info(key, ": ", value)
			err := UpdateCorralConfig(key, value)
			if err != nil {
				errStrings = append(errStrings, fmt.Sprint(err.Error(), "Unable to update corral: "+currentCommand.Name+" for "+key+": "+value))
				break
			}
		}

		cmdToRun, err := startCorral(ts, currentCommand.Name, currentCommand.PackageName, debug, cleanup)
		if err != nil {
			errStrings = append(errStrings, err.Error())
			break
		}

		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			msg, err := runAndWaitOnCommand(cmdToRun)
			if err != nil {
				errStrings = append(errStrings, err.Error())
			}

			msgs = append(msgs, msg)
		}()

	}

	waitGroup.Wait()

	var formattedError error
	var longString string
	if len(errStrings) > 0 {
		for _, err := range errStrings {
			longString += err
		}
		formattedError = errors.New(longString)
	}

	logrus.Info("done with registration")
	return msgs, formattedError
}

// DeleteCorral deletes a corral based on the corral name
func DeleteCorral(corralName string) error {
	msg, err := exec.Command("corral", "delete", corralName).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "DeleteCorral: "+string(msg))
	}
	return nil
}

// ListCorral lists the corrals that currently created
func ListCorral() (map[string]string, error) {
	corralMapList := make(map[string]string)
	msg, err := exec.Command("corral", "list").CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "ListCorral: "+string(msg))
	}
	// The corral list command comes in this format. So the regular
	// expression is we can get the corral name and its state from the commandline
	// +------------+--------+
	// | NAME       | STATUS |
	// +------------+--------+
	// | corralname | READY  |
	// +------------+--------+
	corralNameRegEx := regexp.MustCompile(`\w+ +\| +\w+`)
	corralList := corralNameRegEx.FindAllString(string(msg), -1)
	if len(corralList) == 1 {
		return corralMapList, nil
	}
	for _, corral := range corralList[1:] {
		corralRegEx := regexp.MustCompile(` +\|`)
		corralNameStatus := corralRegEx.Split(corral, -1)
		corralMapList[corralNameStatus[0]] = corralNameStatus[1]
	}
	return corralMapList, nil
}

// GetKubeConfig gets the kubeconfig of corral's cluster
func GetKubeConfig(corral string) ([]byte, error) {
	firstCommand := exec.Command("corral", "vars", corral, "kubeconfig")
	secondCommand := exec.Command("base64", "--decode")

	reader, writer := io.Pipe()
	firstCommand.Stdout = writer
	secondCommand.Stdin = reader

	var byteBuffer bytes.Buffer
	secondCommand.Stdout = &byteBuffer

	err := firstCommand.Start()
	if err != nil {
		return nil, err
	}

	err = secondCommand.Start()
	if err != nil {
		return nil, err
	}

	err = firstCommand.Wait()
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	err = secondCommand.Wait()
	if err != nil {
		return nil, err
	}

	return byteBuffer.Bytes(), nil
}

// UpdateCorralConfig updates a specific corral config var
func UpdateCorralConfig(configVar, value string) error {
	msg, err := exec.Command("corral", "config", "vars", "set", configVar, value).CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "SetupCorralConfig: "+string(msg))
	}
	return nil
}

func DeleteAllCorrals() error {
	corralList, err := ListCorral()
	if err != nil {
		return err
	}
	for corralName := range corralList {
		err := DeleteCorral(corralName)
		logrus.Infof("The Corral %s was deleted.", corralName)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetCorralSSHKeys is a helper function that will set the corral ssh keys previously generated by `corralName`
func SetCorralSSHKeys(corralName string) error {
	privateSSHkey, err := GetCorralEnvVar(corralName, corralPrivateSSHKey)
	if err != nil {
		return err
	}

	privateSSHkey = strings.Replace(privateSSHkey, "\n", "\\n", -1)
	privateSSHkey = fmt.Sprintf("\"%s\"", privateSSHkey)

	err = UpdateCorralConfig(corralPrivateSSHKey, privateSSHkey)
	if err != nil {
		return err
	}

	publicSSHkey, err := GetCorralEnvVar(corralName, corralPublicSSHKey)
	if err != nil {
		return err
	}

	return UpdateCorralConfig(corralPublicSSHKey, publicSSHkey)
}

// SetCorralBastion is a helper function that will set the corral bastion private and pulic addresses previously generated by `corralName`
func SetCorralBastion(corralName string) error {
	bastion_ip, err := GetCorralEnvVar(corralName, corralRegistryIP)
	if err != nil {
		return err
	}

	err = UpdateCorralConfig(corralRegistryIP, bastion_ip)
	if err != nil {
		return err
	}

	bastion_internal_ip, err := GetCorralEnvVar(corralName, corralRegistryPrivateIP)
	if err != nil {
		return err
	}

	return UpdateCorralConfig(corralRegistryPrivateIP, bastion_internal_ip)
}

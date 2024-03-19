package helm

import (
	"net/url"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/rancher/shepherd/pkg/session"
	"github.com/sirupsen/logrus"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	helmAction "helm.sh/helm/v3/pkg/action"
	helmChart "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
)

// A representation of shepherd's helm client which wraps a subset of functionality
// from the upstream helm/v3 packages. Namely the cli, action, release and repo packages.
// Contains a custom RESTClientGetter, helm EnvSettings struct, helm Action Configuration struct
// and bundles a shepherd Session struct if desired
type Client struct {
	RESTGetter    genericclioptions.RESTClientGetter
	Settings      *helmCLI.EnvSettings
	Configuration *helmAction.Configuration
	ts            *session.Session
}

// NewClient initializes a helm Action Configuration struct using either the provided RESTClientGetter or helm EnvSettings
// returns  a new instance of the helm Client and/or returns an error
func NewClient(ts *session.Session, settings *helmCLI.EnvSettings, getter *RESTClientGetter, namespace, helmDriver string) (*Client, error) {
	var actionConfig *helmAction.Configuration
	var err error
	c := &Client{ts: ts}
	if settings != nil {
		if actionConfig, err = InitActionConfig(settings, namespace, helmDriver); err != nil {
			return nil, errors.Wrap(err, "Unable to create new Helm Action Configuration via Helm EnvSettings.")
		}
		c.Settings = settings
		c.RESTGetter = settings.RESTClientGetter()
		c.Configuration = actionConfig
	}
	if getter != nil {
		if actionConfig, err = InitActionConfigWithGetter(getter, namespace, helmDriver); err != nil {
			return nil, errors.Wrap(err, "Unable to create new Helm Action Configuration via RESTClientGetter")
		}
		if c.Settings == nil {
			c.Settings = helmCLI.New()
			if namespace != "" {
				c.Settings.SetNamespace(namespace)
			}
		}
		c.RESTGetter = getter
		c.Configuration = actionConfig
	}
	return c, nil
}

// GetChart is a convenience function that looks for a chart dir before loading it
// Returns a helm Chart struct or an error if any
func GetChart(opts helmAction.ChartPathOptions, chartName string, s *helmCLI.EnvSettings) (*helmChart.Chart, error) {
	chartPath, err := opts.LocateChart(chartName, s)
	if err != nil {
		return nil, errors.Wrapf(err, "GetChart: ")
	}
	chart, err := loader.Load(chartPath)
	if err != nil {
		return nil, errors.Wrapf(err, "GetChart: ")
	}
	return chart, nil
}

// InstallChart installs a helm chart using helm v3 sdk.
// Registers a cleanup function by default, if the test Session is not nil
// Send the helm set command strings such as "--set", "installCRDs=true"
// in the args argument to be prepended to the helm install command.
// returns a helm Release struct and an error if any
// Example of equivalent helm cli command:
//
//	Given: helmClient.InstallChart("rancher", "rancher-latest/rancher", "https://releases.rancher.com/server-charts/latest", "cattle-system", "2.7.10")
//	CLI equivalent: helm install rancher rancher-latest/rancher --repo https://releases.rancher.com/server-charts/latest --namespace cattle-system --version 2.7.10
func (c *Client) InstallChart(releaseName, repoName, repoURL, namespace, version string, cleanup bool, vals map[string]interface{}) (*release.Release, error) {
	var err error
	// Register cleanup function
	if c.ts != nil && cleanup {
		c.ts.RegisterCleanupFunc(func() error {
			uninstallRelease, err := c.UninstallChart(releaseName, true)
			return errors.Wrapf(err, "InstallChart (Cleanup): Release.Info = {%s}", uninstallRelease.Info)

		})
	}

	installClient := helmAction.NewInstall(c.Configuration)
	installClient.Namespace = namespace
	installClient.ReleaseName = releaseName
	installClient.RepoURL = repoURL
	installClient.Version = version
	if repoURL == "" {
		repoURL, err = GetRepoURL(c.Settings.RepositoryConfig, repoName)
		if err != nil {
			return nil, errors.Wrapf(err, "InstallChart: ")
		}
	}

	chartURL, err := repo.FindChartInRepoURL(repoURL, releaseName,
		installClient.Version, installClient.CertFile, installClient.KeyFile,
		installClient.CaFile, getter.All(c.Settings),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "InstallChart: ")
	}

	chart, err := GetChart(installClient.ChartPathOptions, chartURL, c.Settings)
	if err != nil {
		return nil, errors.Wrapf(err, "InstallChart: ")
	}
	return installClient.Run(chart, vals)
}

// InstallChartWithOptions installs a helm chart using helm v3 sdk.
// Registers a cleanup function by default, if the test Session is not nil
// Send the helm set command strings such as "--set", "installCRDs=true"
// in the args argument to be prepended to the helm install command.
// returns a helm Release struct and an error if any
// Example of equivalent helm cli command:
//
//	Given: helmClient.InstallChart("rancher", "rancher-latest/rancher", "https://releases.rancher.com/server-charts/latest", "cattle-system", "2.7.10")
//	CLI equivalent: helm install rancher rancher-latest/rancher --repo https://releases.rancher.com/server-charts/latest --namespace cattle-system --version 2.7.10
func (c *Client) InstallChartWithOptions(releaseName, repoName string, cleanup bool, vals map[string]interface{}, opts *helmAction.Install) (*release.Release, error) {
	// Register cleanup function
	if opts != nil && opts.DryRun {
		cleanup = false
	}
	if c.ts != nil && cleanup {
		c.ts.RegisterCleanupFunc(func() error {
			uninstallRelease, err := c.UninstallChart(releaseName, true)
			return errors.Wrapf(err, "InstallChartWithOptions (Cleanup): Release.Info = {%s}", uninstallRelease.Info)

		})
	}

	installClient := helmAction.NewInstall(c.Configuration)
	err := mergo.Merge(installClient, opts, mergo.WithOverride, mergo.WithoutDereference)
	if err != nil {
		return nil, errors.Wrapf(err, "InstallChartWithOptions: ")
	}

	repoURL, err := GetRepoURL(c.Settings.RepositoryConfig, repoName)
	if err != nil {
		return nil, errors.Wrapf(err, "InstallChartWithOptions: ")
	}

	chartURL, err := repo.FindChartInRepoURL(repoURL, releaseName,
		installClient.Version, installClient.CertFile, installClient.KeyFile,
		installClient.CaFile, getter.All(c.Settings),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "InstallChartWithOptions: ")
	}

	chart, err := GetChart(installClient.ChartPathOptions, chartURL, c.Settings)
	if err != nil {
		logrus.Info(err)
		return nil, errors.Wrapf(err, "InstallChartWithOptions: ")
	}
	return installClient.Run(chart, vals)
}

// UpgradeChart is a convenience function for upgrading a helm chart using helm v3 sdk.
// Allows for a limited set of the equivalent cli arguments to be passed as parameters.
// Registers a cleanup function by default, if the test Session is not nil.
// returns a helm Release struct and an error if any
// In order to construct a `helm upgrade` command with fields that are not set in this function,
// you must construct a `*helmAction.Upgrade` directly, without using this function.
// Example of equivalent helm cli command:
//
//	Given: helmClient.UpgradeChart("rancher", "rancher-latest/rancher", "https://releases.rancher.com/server-charts/latest", "cattle-system", "2.7.10", true, vals)
//	CLI equivalent: helm upgrade rancher rancher-latest/rancher --repo https://releases.rancher.com/server-charts/latest --namespace cattle-system --version 2.7.10 [--set <each value in vals>]
func (c *Client) UpgradeChart(releaseName, repoName, repoURL, namespace, version string, devel, cleanup bool, vals map[string]interface{}) (*release.Release, error) {
	var err error
	// Register cleanup function
	if c.ts != nil && cleanup {
		c.ts.RegisterCleanupFunc(func() error {
			uninstallRelease, err := c.UninstallChart(releaseName, true)
			return errors.Wrapf(err, "UpgradeChart (Cleanup): Release.Info = {%s}", uninstallRelease.Info)
		})
	}

	upgradeClient := helmAction.NewUpgrade(c.Configuration)
	upgradeClient.Namespace = namespace
	upgradeClient.RepoURL = repoURL
	upgradeClient.Devel = devel
	upgradeClient.Version = version
	if repoURL == "" {
		upgradeClient.RepoURL, err = GetRepoURL(c.Settings.RepositoryConfig, repoName)
		if err != nil {
			return nil, errors.Wrapf(err, "UpgradeChart: ")
		}
	}

	chartReference := repoName + "/" + releaseName
	chart, err := GetChart(upgradeClient.ChartPathOptions, chartReference, c.Settings)
	if err != nil {
		return nil, errors.Wrapf(err, "UpgradeChart: ")
	}

	return upgradeClient.Run(releaseName, chart, vals)
}

// UpgradeChart is a convenience function for upgrading a helm chart using helm v3 sdk.
// Allows for a limited set of the equivalent cli arguments to be passed as parameters.
// Registers a cleanup function by default, if the test Session is not nil.
// Returns a helm Release struct and an error if any.
// In order to construct a `helm upgrade` command with fields that are not set in this function,
// you must construct a `*helmAction.Upgrade` directly, without using this function.
// Example of equivalent helm cli command:
//
//	Given: helmClient.UpgradeChartWithOptions("rancher", "rancher-latest", vals, opts)
//	CLI equivalent: helm upgrade rancher rancher-latest/rancher [--set <each value in vals>] --[Additional options here]
func (c *Client) UpgradeChartWithOptions(releaseName, repoName string, cleanup bool, vals map[string]interface{}, opts *helmAction.Upgrade) (*release.Release, error) {
	// Register cleanup function
	if opts != nil && opts.DryRun {
		cleanup = false
	}
	if c.ts != nil && cleanup {
		c.ts.RegisterCleanupFunc(func() error {
			uninstallRelease, err := c.UninstallChart(releaseName, true)
			if uninstallRelease != nil {
				return errors.Wrapf(err, "UpgradeChartWithOptions (Cleanup): Release.Info = {%s}", uninstallRelease.Info)
			}
			return err
		})
	}

	upgradeClient := helmAction.NewUpgrade(c.Configuration)
	err := mergo.Merge(upgradeClient, opts, mergo.WithOverride, mergo.WithoutDereference)
	if err != nil {
		return nil, errors.Wrapf(err, "UpgradeChartWithOptions: ")
	}

	upgradeClient.RepoURL, err = GetRepoURL(c.Settings.RepositoryConfig, repoName)
	if err != nil {
		return nil, errors.Wrapf(err, "UpgradeChartWithOptions: ")
	}
	err = c.AddOrUpdateRepo(repoName, upgradeClient.RepoURL, cleanup)
	if err != nil {
		return nil, errors.Wrapf(err, "UpgradeChartWithOptions: ")
	}

	chart, err := GetChart(upgradeClient.ChartPathOptions, releaseName, c.Settings)
	if err != nil {
		return nil, errors.Wrapf(err, "UpgradeChartWithOptions: ")
	}

	return upgradeClient.Run(releaseName, chart, vals)
}

func (c *Client) GetValues(releaseName string) (map[string]interface{}, error) {
	return helmAction.NewGetValues(c.Configuration).Run(releaseName)
}

// UninstallChart uninstalls a helm chart using helm CLI in a given namespace
// using the releaseName provided.
// returns an helm UninstallReleaseResponse struct or error if any.
func (c *Client) UninstallChart(releaseName string, wait bool) (*release.UninstallReleaseResponse, error) {
	// Default helm uninstall command
	uninstallClient := helmAction.NewUninstall(c.Configuration)
	uninstallClient.Wait = wait
	return uninstallClient.Run(releaseName)
}

// UninstallChart uninstalls a helm chart using helm CLI in a given namespace
// using the releaseName provided.
// Returns an helm UninstallReleaseResponse struct or error if any.
func (c *Client) UninstallChartWithOptions(releaseName string, opts *helmAction.Uninstall) (*release.UninstallReleaseResponse, error) {
	uninstallClient := helmAction.NewUninstall(c.Configuration)
	err := mergo.Merge(uninstallClient, opts, mergo.WithOverride, mergo.WithoutDereference)
	if err != nil {
		return nil, errors.Wrapf(err, "UninstallChartWithOptions: ")
	}
	return uninstallClient.Run(releaseName)
}

// GetRepoFile is a convenience function that handles pulling the local helm repository
// config file which is needed to easily reference existing helm repository information.
// Takes in a string path to the repository config file to use.
// Returns the *repo.File and an error if any.
func GetRepoFile(repositoryConfig string) (*repo.File, error) {
	repoFile, err := repo.LoadFile(repositoryConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "GetRepoFile: ")
	}
	return repoFile, err
}

// GetRepoURL is a convenience function that handles extracting the repository URL
// for pre-existing repositories that have been added to the local machine's
// helm repository config file.
// Takes in 2 strings: the path to the helm repository config file and a repository name.
// Useful for finding charts stored in remote repositories (combine this with helm.sh/helm/v3/pkg/repo.FindChartInRepoURL()).
// Returns the repository URL and an error if any.
func GetRepoURL(repositoryConfig, repoName string) (string, error) {
	repoFile, err := GetRepoFile(repositoryConfig)
	if err != nil {
		return "", errors.Wrapf(err, "GetRepoURL: ")
	}

	repoURL := ""
	if !repoFile.Has(repoName) {
		return "", errors.Wrapf(err, "GetRepoURL: ")
	}

	repoURL = repoFile.Get(repoName).URL

	return repoURL, nil
}

// AddOrUpdateRepo adds or updates the specified helm repository if it already exists
// Does not allow 0-length strings as input, repoURL must be a valid fully-qualified URL
// Returns an error if any
func (c *Client) AddOrUpdateRepo(repoName, repoURL string, cleanup bool) error {
	if repoName == "" || repoURL == "" {
		return errors.New("func AddOrUpdateRepo: repository name and URL must be strings with length > 0")
	}
	if _, err := url.ParseRequestURI(repoURL); err != nil {
		return errors.Wrapf(err, "AddorUpdateRepo: repository URL must be valid fully-qualified URL")
	}
	repoFile, err := GetRepoFile(c.Settings.RepositoryConfig)
	if err != nil {
		return errors.Wrapf(err, "AddOrUpdateRepo: ")
	}
	repoEntry := &repo.Entry{
		Name: repoName,
		URL:  repoURL,
	}

	repoFile.Update(repoEntry)
	err = repoFile.WriteFile(c.Settings.RepositoryConfig, 0755)
	if err != nil {
		return errors.Wrapf(err, "AddOrUpdateRepo: ")
	}

	// Register cleanup function
	if c.ts != nil && cleanup {
		c.ts.RegisterCleanupFunc(func() error {
			if repoFile.Has(repoName) {
				removed, err := c.RemoveRepo(repoName)
				if err != nil {
					return errors.Wrapf(err, "AddOrUpdateRepo: ")
				}
				if !removed {
					return errors.Errorf("AddOrUpdateRepo: Failed to remove repository named '%s'", repoName)
				}
			} else {
				return errors.Errorf("AddOrUpdateRepo: Failed to add repository named '%s'", repoName)
			}
			return nil
		})
	}
	return nil
}

// RemoveRepo removes the specified helm repository from the local repository config file, if it exists
// Returns a bool representing if the repository was removed or not and an error if any
func (c *Client) RemoveRepo(repoName string) (bool, error) {
	repoFile, err := GetRepoFile(c.Settings.RepositoryConfig)
	if err != nil {
		return false, errors.Wrapf(err, "RemoveRepo: ")
	}

	removed := repoFile.Remove(repoName)
	err = repoFile.WriteFile(c.Settings.RepositoryConfig, 0755)
	if err != nil {
		return false, errors.Wrapf(err, "RemoveRepo: ")
	}

	return removed, nil
}

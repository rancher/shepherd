package charts

import (
	"context"
	"errors"

	v1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	"github.com/rancher/shepherd/clients/rancher/catalog"
	"github.com/rancher/shepherd/extensions/defaults"
	"github.com/rancher/shepherd/pkg/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// WaitChartInstall verifies that the app from a chart was successfully deployed
func WaitChartInstall(client *catalog.Client, chartNamespace, chartName string) error {
	watchAppInterface, err := client.Apps(chartNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + chartName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*v1.App)

		state := app.Status.Summary.State
		if state == string(v1.StatusDeployed) {
			return true, nil
		}

		if state == string(v1.StatusFailed) {
			return false, errors.New("chart install has failed")
		}
		return false, nil
	})
	return err
}

// WaitChartUpgrade verifies that the app from a chart was successfully upgraded
func WaitChartUpgrade(client *catalog.Client, chartNamespace, chartName, chartVersion string) error {
	watchAppInterface, err := client.Apps(chartNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + chartName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*v1.App)

		state := app.Status.Summary.State
		if state == string(v1.StatusPendingUpgrade) {
			return true, nil
		}

		if state == string(v1.StatusFailed) {
			return false, errors.New("chart upgrade has failed")
		}
		return false, nil
	})
	if err != nil {
		return err
	}

	watchAppInterface, err = client.Apps(chartNamespace).Watch(context.TODO(), metav1.ListOptions{
		FieldSelector:  "metadata.name=" + chartName,
		TimeoutSeconds: &defaults.WatchTimeoutSeconds,
	})
	if err != nil {
		return err
	}

	err = wait.WatchWait(watchAppInterface, func(event watch.Event) (ready bool, err error) {
		app := event.Object.(*v1.App)

		state := app.Status.Summary.State
		if state == string(v1.StatusDeployed) {
			if app.Spec.Chart.Metadata.Version == chartVersion {
				return true, nil
			}
			return false, errors.New("chart not upgraded to expected version: " + chartVersion)
		}

		if state == string(v1.StatusFailed) {
			return false, errors.New("chart upgrade has failed")
		}
		return false, nil
	})
	return err
}

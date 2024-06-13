package wait

import (
	"github.com/rancher/shepherd/pkg/wait"
	"k8s.io/apimachinery/pkg/watch"
)

// ResourceCreate is a generic wait function for create operations on v1 resources
func ResourceCreate(watchInterface watch.Interface) error {
	err := wait.WatchWait(watchInterface, func(event watch.Event) (ready bool, err error) {
		if event.Type == watch.Added {
			return true, nil
		} else if event.Type == watch.Error {
			return false, nil
		}

		return false, nil
	})

	return err

}

// ResourceCreate is a generic wait function for delete operations on v1 resources
func ResourceDelete(watchInterface watch.Interface) error {
	err := wait.WatchWait(watchInterface, func(event watch.Event) (ready bool, err error) {
		if event.Type == watch.Error {
			return false, nil
		} else if event.Type == watch.Deleted {
			return true, nil
		}

		return false, nil
	})

	return err
}

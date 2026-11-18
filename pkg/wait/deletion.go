package wait

import (
	"context"
	"fmt"
	"time"

	"github.com/rancher/shepherd/extensions/defaults"
	"k8s.io/apimachinery/pkg/api/errors"
	kwait "k8s.io/apimachinery/pkg/util/wait"
)

// ForDeletion polls get until it reports the object NotFound or Forbidden, or returns why the deletion could not be confirmed
func ForDeletion(get func(context.Context) error, timeout time.Duration) error {
	var lastGetErr error

	waitErr := kwait.PollUntilContextTimeout(context.TODO(), defaults.FiveHundredMillisecondTimeout, timeout, true, func(ctx context.Context) (bool, error) {
		getErr := get(ctx)

		if errors.IsNotFound(getErr) || errors.IsForbidden(getErr) {
			return true, nil
		}

		lastGetErr = getErr

		return false, nil
	})
	if waitErr != nil && lastGetErr != nil {
		return fmt.Errorf("could not be read back to confirm it is gone (last read error: %w)", lastGetErr)
	}

	if waitErr != nil {
		return fmt.Errorf("is still present after %v", timeout)
	}

	return nil
}

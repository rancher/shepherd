package kubectl

import (
	"context"
	"fmt"
	"strings"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/strings/slices"

	v1Unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TODO: attempt to create generic updatefields func
func UpdateFields(obj *v1Unstructured.Unstructured, fields map[string]interface{}) error {

	for k, v := range fields {
		if strings.Contains(k, ".") {
			// Maybe replace with Split()
			keyFields := strings.Split(k, ".")
			// curr, nestedFields, found := strings.Cut(k, ".")
			// if !found || len(nestedFields) == 0 || len(curr) == 0 {
			// 	panic(fmt.Errorf("nested fields not found or error in field name: %s", k))
			// }
			if slices.Contains(keyFields, "") {
				panic(fmt.Errorf("error in field name, found \"\" in: %s", k))
			}
			nestedSlice, nestedFound, err := v1Unstructured.NestedSlice(obj.Object, keyFields...)
			if err != nil || !nestedFound || nestedSlice == nil {
				fmt.Errorf("nested slice not found or error in spec: %v", err)
			} else {
				return v1Unstructured.SetNestedField(nestedSlice[0].(map[string]interface{}), v, keyFields[1:]...)
			}
			nestedMap, nestedFound, err := v1Unstructured.NestedMap(obj.Object, keyFields...)
			if err != nil || !nestedFound || nestedMap == nil {
				fmt.Errorf("nested map not found or error in spec: %v", err)
			}
		} else {
			v1Unstructured.SetNestedField(obj.Object, v, k)
		}
	}
	return nil
}

func UpdateUnstructured(s *session.Session, client *rancher.Client, obj *v1Unstructured.Unstructured, fields map[string]interface{}, clusterID, n string, gvr schema.GroupVersionResource) (*v1Unstructured.Unstructured, error) {
	dynClient, _, err := setupDynamicClient(s, client, nil, clusterID)
	if err != nil {
		return nil, err
	}
	var result *v1Unstructured.Unstructured
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		// update replicas to 1
		err := UpdateFields(obj, fields)
		if err != nil {
			return err
		}
		result, err = dynClient.Resource(gvr).Namespace(n).Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
	if retryErr != nil {
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
	return result, nil
}

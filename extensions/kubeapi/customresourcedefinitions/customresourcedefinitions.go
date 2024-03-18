package customresourcedefinitions

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// gets a list of names of custom resource definitions that contain the input string name from an Unstructured List
func GetCustomResourceDefinitionsListByName(CRDList *unstructured.UnstructuredList, name string) []string {
	var CRDNameList []string
	CRDs := *CRDList
	for _, unstructuredCRD := range CRDs.Items {
		CRDName := unstructuredCRD.GetName()
		if strings.Contains(CRDName, name) {
			CRDNameList = append(CRDNameList, CRDName)
		}
	}

	return CRDNameList
}

package utils

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewUnstructuredList(items ...unstructured.Unstructured) *unstructured.UnstructuredList {
	res := &unstructured.UnstructuredList{
		Items: []unstructured.Unstructured{},
	}
	res.Items = append(res.Items, items...)
	return res
}

package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PatchAnnotations patches the object to ensure that the specified annotations exist.
// Does not patch if all annotations already have their expected values.
func (k *Kubeconfig) PatchAnnotations(ctx context.Context, obj client.Object, anns map[string]string) (bool, error) {
	oldAnns := obj.GetAnnotations()
	if oldAnns != nil {
		allSet := true
		for k, v := range anns {
			if oldV, ok := oldAnns[k]; !ok || oldV != v {
				allSet = false
				break
			}
		}
		if allSet {
			// all annotations already have the correct values, nothing to do
			return false, nil
		}
	}
	patchData, err := json.Marshal(anns)
	if err != nil {
		return false, err
	}
	return true, k.Patch(ctx, obj, client.RawPatch(types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"annotations":%s}}`, string(patchData)))))
}

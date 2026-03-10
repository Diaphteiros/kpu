package utils

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetFinalizerIndices takes an k8s object and a list of finalizers and returns the indices of the corresponding entries in the object's finalizers.
// If a given finalizer doesn't exist on the object, not index is returned for it.
// This means that from the resulting list of indices, it cannot be reconstructed which finalizer is at a specific index, just that it is one of the given ones.
func GetFinalizerIndices(obj client.Object, fins ...string) []int {
	res := []int{}
	for i, f := range obj.GetFinalizers() {
		if slices.Contains(fins, f) {
			res = append(res, i)
		}
	}
	return res
}

// RemoveFinalizers removes all instances of all specified finalizers from the given object and patches it in the cluster.
// If fins is empty, all finalizers are removed.
func (k *Kubeconfig) RemoveFinalizers(ctx context.Context, obj client.Object, fins ...string) (bool, error) {
	var patch string
	if len(obj.GetFinalizers()) == 0 {
		return false, nil
	}
	if len(fins) == 0 {
		patch = "[{\"op\": \"remove\", \"path\": \"/metadata/finalizers\"}]"
	} else {
		fis := GetFinalizerIndices(obj, fins...)
		if len(fis) == 0 {
			return false, nil
		}
		patchb := strings.Builder{}
		patchb.WriteString("[")
		for i, fidx := range fis {
			patchb.WriteString("{\"op\": \"remove\", \"path\": \"/metadata/finalizers/")
			patchb.WriteString(fmt.Sprint(fidx))
			patchb.WriteString("\"}")
			if i < len(fis)-1 {
				patchb.WriteString(", ")
			}
		}
		patchb.WriteString("]")
		patch = patchb.String()
	}
	return true, k.Patch(ctx, obj, client.RawPatch(types.JSONPatchType, []byte(patch)))
}

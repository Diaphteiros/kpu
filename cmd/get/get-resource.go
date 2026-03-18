package get

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"

	"github.com/Diaphteiros/kpu/pkg/utils"
)

// GetResourceCmd represents the 'get resource' command
var GetResourceCmd = &cobra.Command{
	Use:     "resource",
	Aliases: []string{"res", "r"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Get the specified k8s resources",
	Long: `This is pretty similar to 'kubectl get' and mainly in here for debugging purposes.

Examples:

	> kpu get resource namespace foo -o yaml
	Returns the namespace 'foo' in yaml format.

	> kpu get resource pod,deployment -n foo -l example.org/mylabel
	Lists all pods and deployments in the namespace 'foo' with the label 'example.org/mylabel' (independent of the label value).
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateGetResourceCommand(args)

		resourceTypes := strings.Split(args[0], ",")
		var resourceNames []string
		if len(args) > 1 {
			resourceNames = []string{} //nolint:prealloc
			for _, arg := range args[1:] {
				resourceNames = append(resourceNames, strings.Split(arg, ",")...)
			}
			slices.Filter(resourceNames[:0], resourceNames, func(s string) bool { return s != "" })
		}

		k, err := utils.LoadKubeconfigWithImpersonation(k8sOptions.KubeconfigPath, k8sOptions.ImpersonationConfig)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		objects := utils.NewUnstructuredList()
		var errs []error
		objects.Items, errs = k.ListResources(cmd.Context(), resourceTypes, resourceNames, utils.SCOPE_ALL, k8sOptions)

		if err := errors.Join(errs...); err != nil {
			utils.Fatal(1, "error getting resources: %s", err.Error())
		}

		if !showManagedFields && output != utils.OUTPUT_TEXT {
			// filter out managedFields
			for i := range objects.Items {
				obj := &objects.Items[i]
				obj.SetManagedFields(nil)
			}
		}

		printUnstructuredObjects(output, objects)
	},
}

func init() {
	GetResourceCmd.Flags().BoolVar(&showManagedFields, "show-managed-fields", false, "If true, keep the managedFields when printing objects in JSON or YAML format.")
	utils.AddOutputFlag(GetResourceCmd.Flags(), &output, utils.OUTPUT_TEXT)
}

func ValidateGetResourceCommand(args []string) {
	utils.ValidateOutputFormat(output)
	if len(args) == 0 {
		utils.Fatal(1, "no resource kinds specified, use 'kpu get all' to get all resources")
	}
}

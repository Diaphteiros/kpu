package get

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Diaphteiros/kpu/pkg/utils"
)

// variables for holding the flags
var (
	suppressWarnings  bool
	output            utils.OutputFormat
	scope             utils.Scope
	showManagedFields bool
)

// GetAllCmd represents the 'get all' command
var GetAllCmd = &cobra.Command{
	Use:     "all",
	Aliases: []string{"a"},
	Args:    cobra.NoArgs,
	Short:   "Get all k8s resources",
	Long: `This command does what 'kubectl get all' is supposed to do:
It gets ALL k8s resources.

Use the --scope flag to control if you want to get namespace- or cluster-scoped resources (or both).

Examples:

	> kpu get all
	Lists all namespace-scoped resources in the default namespace.

	> kpu get all --scope cluster
	Lists all cluster-scoped resources.
`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateGetAllCommand()

		k, err := utils.LoadKubeconfigWithImpersonation(k8sOptions.KubeconfigPath, k8sOptions.ImpersonationConfig)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		objects := utils.NewUnstructuredList()
		var errs []error
		objects.Items, errs = k.GetAllResources(cmd.Context(), scope, k8sOptions)

		if !suppressWarnings && len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "warning: %s\n", e.Error())
			}
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
	GetAllCmd.Flags().BoolVarP(&suppressWarnings, "suppress-warnings", "s", false, "If true, no warnings will be printed to stderr if objects of a specific kind could not be listed.")
	GetAllCmd.Flags().BoolVar(&showManagedFields, "show-managed-fields", false, "If true, keep the managedFields when printing objects in JSON or YAML format.")
	utils.AddScopeFlag(GetAllCmd.Flags(), &scope)
	utils.AddOutputFlag(GetAllCmd.Flags(), &output, utils.OUTPUT_TEXT)
}

func ValidateGetAllCommand() {
	utils.ValidateOutputFormat(output)
	if scope == utils.SCOPE_CLUSTER && (k8sOptions.Namespace != "" || k8sOptions.AllNamespaces) {
		utils.Fatal(1, "the --scope=cluster is incompatible with --namespace and --all-namespaces")
	}
}

package del

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
)

// variables for holding the flags
var (
	k8sOptions       *utils.K8sInteractionOptions = &utils.K8sInteractionOptions{}
	autoConfirmation bool
	allResources     bool
	suppressWarnings bool
)

// DeleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:     "delete type1[,type2,...] [name1,name2 [name3 ...]]",
	Aliases: []string{"del", "d"},
	Args:    cobra.MinimumNArgs(1),
	GroupID: cmdgroups.ClusterInteraction,
	Short:   "Delete k8s resources",
	Long: `This command works basically like 'kubectl delete', with two major differences:
1. It lists all affected resources and asks for confirmation beforehand (unless -y is specified).
2. It adds deletion confirmation annotations for specific API groups before attempting the deletion.

Currently, the following deletion confirmation rules are implemented:
- core.gardener.cloud => confirmation.gardener.cloud/deletion=true
- openmcp.cloud       => confirmation.openmcp.cloud/deletion=true
`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateDeleteCommand(args)

		k, err := utils.LoadKubeconfig(k8sOptions.KubeconfigPath)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		// split first argument by ','
		resourceTypes := strings.Split(args[0], ",")

		var resourceNames []string
		if len(args) > 1 {
			resourceNames = []string{} //nolint:prealloc
			for _, arg := range args[1:] {
				resourceNames = append(resourceNames, strings.Split(arg, ",")...)
			}
			slices.Filter(resourceNames[:0], resourceNames, func(s string) bool { return s != "" })
		}

		affectedResources, errs := k.ListResources(cmd.Context(), resourceTypes, resourceNames, utils.SCOPE_ALL, k8sOptions)

		if !suppressWarnings {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "warning: %s\n", e.Error())
			}
		}
		if len(affectedResources) == 0 {
			fmt.Println("No resources would be affected by this command.")
			os.Exit(0)
		}

		if !autoConfirmation {
			// build prompt
			p := strings.Builder{}
			p.WriteString("This will delete the following ")
			if len(affectedResources) == 1 {
				p.WriteString("resource:\n")
			} else {
				fmt.Fprint(&p, len(affectedResources))
				p.WriteString(" resources:\n")
			}
			for _, obj := range affectedResources {
				p.WriteString(utils.GVKToString(obj.GroupVersionKind()))
				p.WriteString("\t")
				ns := obj.GetNamespace()
				if ns != "" {
					p.WriteString(ns)
					p.WriteString("/")
				}
				p.WriteString(obj.GetName())
				p.WriteString("\n")
			}
			p.WriteString("Do you want to continue?")
			if !utils.PromptForConfirmation(p.String(), false) {
				fmt.Println("Command aborted.")
				os.Exit(0)
			}
		}

		errOccurred := false
		for _, obj := range affectedResources {
			var anns map[string]string
			switch obj.GroupVersionKind().Group {
			case "core.gardener.cloud":
				anns = map[string]string{
					"confirmation.gardener.cloud/deletion": "true",
				}
			case "core.openmcp.cloud":
				anns = map[string]string{
					"confirmation.openmcp.cloud/deletion": "true",
				}
			}
			if len(anns) > 0 {
				changed, err := k.PatchAnnotations(cmd.Context(), &obj, anns)
				if err != nil {
					fmt.Printf("error patching %s: %s\n", utils.ResourceIdentifier(&obj), err.Error())
					errOccurred = true
				} else {
					if changed {
						fmt.Printf("%s patched\n", utils.ResourceIdentifier(&obj))
					} else {
						fmt.Printf("%s unchanged\n", utils.ResourceIdentifier(&obj))
					}
				}
			}
			if err := k.Delete(cmd.Context(), &obj); err != nil {
				fmt.Printf("error deleting %s: %s\n", utils.ResourceIdentifier(&obj), err.Error())
				errOccurred = true
			}
		}
		if errOccurred {
			os.Exit(1)
		}
	},
}

func init() {
	utils.AddDefaultK8sInteractionFlags(DeleteCmd.PersistentFlags(), k8sOptions)
	DeleteCmd.Flags().BoolVar(&allResources, "all", false, "Affect all resources of the chosen type(s).")
	DeleteCmd.Flags().BoolVarP(&autoConfirmation, "yes", "y", false, "If true, command won't prompt for confirmation.")
	DeleteCmd.Flags().BoolVarP(&suppressWarnings, "suppress-warnings", "s", false, "If true, no warnings will be printed to stderr if objects of a specific kind could not be listed or resources were not found.")
}

func ValidateDeleteCommand(args []string) {
	if allResources != (len(args) < 2) {
		utils.Fatal(1, "either resource names or --all have to be specified")
	}
}

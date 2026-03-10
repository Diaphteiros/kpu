package finalize

import (
	"fmt"
	"os"
	"strings"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
	"github.com/spf13/cobra"
	"k8s.io/utils/strings/slices"
)

// variables for holding the flags
var (
	k8sOptions       *utils.K8sInteractionOptions = &utils.K8sInteractionOptions{}
	scope            utils.Scope
	finalizers       []string
	allResources     bool
	suppressWarnings bool
	autoConfirmation bool
)

// FinalizeCmd represents the finalize command
var FinalizeCmd = &cobra.Command{
	Use:     "finalize [type1[,type2,...]] [name1,name2 [name3 ...]]",
	Aliases: []string{"fin", "f"},
	Args:    cobra.MinimumNArgs(1),
	GroupID: cmdgroups.ClusterInteraction,
	Short:   "Remove finalizers from k8s resources",
	Long: `This command can remove all or only specific finalizers from specific or all resources.

Examples:

	> kpu finalize secret foo -n bar
	Removes all finalizers from the secret 'foo' in the namespace 'bar'.
	
	> kpu finalize --all -n bar
	Removes all finalizers from all resources in the namespace 'bar'.
	
	> kpu finalize -f foo.bar.baz/foobar --all -A
	Removes the finalizer 'foo.bar.baz/foobar' from all resources in all namespaces.
	Note that this affects only namespace-scoped resources due to the default value 'namespaced' for --scope.

	> kpu finalize secret,configmap foo bar -f foo.bar.baz/foobar -f foo.bar.baz/asdf
	Removes the finalizers 'foo.bar.baz/foobar' and 'foo.bar.baz/asdf' from the secrets and configmaps
	named 'foo' and 'bar' in the default namespace.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateFinalizeCommand(args)

		k, err := utils.LoadKubeconfig(k8sOptions.KubeconfigPath)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		var resourceTypes []string
		if len(args) > 0 {
			// resource types given
			// split first argument by ','
			resourceTypes = strings.Split(args[0], ",")
		}

		var resourceNames []string
		if len(args) > 1 {
			resourceNames = []string{}
			for _, arg := range args[1:] {
				resourceNames = append(resourceNames, strings.Split(arg, ",")...)
			}
			slices.Filter(resourceNames[:0], resourceNames, func(s string) bool { return s != "" })
		}

		affectedResources, errs := k.ListResources(cmd.Context(), resourceTypes, resourceNames, scope, k8sOptions)

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
			p.WriteString("This will remove all ")
			if len(finalizers) > 0 {
				p.WriteString(utils.NaturalLanguageJoin(finalizers, ", ", true))
				p.WriteString(" ")
			}
			p.WriteString("finalizers from the following ")
			if len(affectedResources) == 1 {
				p.WriteString("resource:\n")
			} else {
				p.WriteString(fmt.Sprint(len(affectedResources)))
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
			changed, err := k.RemoveFinalizers(cmd.Context(), &obj, finalizers...)
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
		if errOccurred {
			os.Exit(1)
		}
	},
}

func init() {
	utils.AddDefaultK8sInteractionFlags(FinalizeCmd.PersistentFlags(), k8sOptions)
	utils.AddScopeFlag(FinalizeCmd.Flags(), &scope)
	FinalizeCmd.Flags().StringSliceVarP(&finalizers, "finalizer", "f", []string{}, "Name(s) of finalizer(s) to be removed. Leave empty to remove all finalizers.")
	FinalizeCmd.Flags().BoolVar(&allResources, "all", false, "Affect all resources of the chosen type(s).")
	FinalizeCmd.Flags().BoolVarP(&autoConfirmation, "yes", "y", false, "If true, command won't prompt for confirmation.")
	FinalizeCmd.Flags().BoolVarP(&suppressWarnings, "suppress-warnings", "s", false, "If true, no warnings will be printed to stderr if objects of a specific kind could not be listed or resources were not found.")
}

func ValidateFinalizeCommand(args []string) {
	if allResources != (len(args) < 2) {
		utils.Fatal(1, "either resource names or --all have to be specified")
	}
}

package reconcile

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
)

// variables for holding the flags
var (
	k8sOptions          *utils.K8sInteractionOptions = &utils.K8sInteractionOptions{}
	allResources        bool
	suppressWarnings    bool
	requireConfirmation bool
	quiet               bool
)

type GroupWithReconcileAnnotation struct {
	Group                    string
	ReconcileAnnotationKey   string
	ReconcileAnnotationValue string
}

var (
	KnownGroupsWithReconcileAnnotation = map[string]*GroupWithReconcileAnnotation{
		"core.gardener.cloud": {
			Group:                    "core.gardener.cloud",
			ReconcileAnnotationKey:   "gardener.cloud/operation",
			ReconcileAnnotationValue: "reconcile",
		},
		"landscaper.gardener.cloud": {
			Group:                    "landscaper.gardener.cloud",
			ReconcileAnnotationKey:   "landscaper.gardener.cloud/operation",
			ReconcileAnnotationValue: "reconcile",
		},
		"openmcp.cloud": {
			Group:                    "openmcp.cloud",
			ReconcileAnnotationKey:   "openmcp.cloud/operation",
			ReconcileAnnotationValue: "reconcile",
		},
		"core.openmcp.cloud": {
			Group:                    "core.openmcp.cloud",
			ReconcileAnnotationKey:   "openmcp.cloud/operation",
			ReconcileAnnotationValue: "reconcile",
		},
	}
	UnknownGroup = &GroupWithReconcileAnnotation{}
)

// ReconcileCmd represents the reconcile command
var ReconcileCmd = &cobra.Command{
	Use:     "reconcile [type1[,type2,...]] [name1,name2 [name3 ...]]",
	Aliases: []string{"rec", "r"},
	Args:    cobra.MinimumNArgs(0),
	GroupID: cmdgroups.ClusterInteraction,
	Short:   "Add reconcile annotations to k8s resources",
	Long: `This command patches reconcile annotations to the specified resources.

Known api groups are:
	core.gardener.cloud       => gardener.cloud/operation: reconcile
	landscaper.gardener.cloud => landscaper.gardener.cloud/operation: reconcile
	openmcp.cloud             => openmcp.cloud/operation: reconcile
	core.openmcp.cloud        => openmcp.cloud/operation: reconcile

Examples:

	> kpu reconcile installation foo -n bar
	Adds the 'landscaper.gardener.cloud/operation: reconcile' annotation to the installation 'foo' in the namespace 'bar'.
	
	> kpu reconcile shoot --all -n garden-myproject
	Adds the 'gardener.cloud/operation: reconcile' annotation to all shoots in the namespace 'garden-myproject'.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateReconcileCommand(args)

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

		knownGroups := sets.KeySet(KnownGroupsWithReconcileAnnotation)
		affectedResourcesPerGroup := map[*GroupWithReconcileAnnotation][]client.Object{}
		affectedResourcesPerGroup[UnknownGroup] = []client.Object{}
		for _, group := range KnownGroupsWithReconcileAnnotation {
			knownGroups.Insert(group.Group)
			affectedResourcesPerGroup[group] = []client.Object{}
		}
		for _, obj := range affectedResources {
			group := obj.GroupVersionKind().Group
			if knownGroups.Has(group) {
				affectedResourcesPerGroup[KnownGroupsWithReconcileAnnotation[group]] = append(affectedResourcesPerGroup[KnownGroupsWithReconcileAnnotation[group]], &obj)
			} else {
				affectedResourcesPerGroup[UnknownGroup] = append(affectedResourcesPerGroup[UnknownGroup], &obj)
			}
		}

		if !suppressWarnings && len(affectedResourcesPerGroup[UnknownGroup]) > 0 {
			w := strings.Builder{}
			affectedObjs := affectedResourcesPerGroup[UnknownGroup]
			w.WriteString("warning: the following ")
			if len(affectedObjs) == 1 {
				w.WriteString("resource")
			} else {
				fmt.Fprint(&w, len(affectedObjs))
				w.WriteString(" resources")
			}
			w.WriteString(" do not belong to one of the known api groups and will not be annotated:\n")
			for _, obj := range affectedObjs {
				w.WriteString("\t")
				w.WriteString(utils.ResourceIdentifier(obj))
				w.WriteString("\n")
			}
			fmt.Fprint(os.Stderr, w.String())
		}
		if requireConfirmation {
			// build prompt
			p := strings.Builder{}
			for _, key := range sets.List(knownGroups) {
				groupWR := KnownGroupsWithReconcileAnnotation[key]
				affectedObjs := affectedResourcesPerGroup[groupWR]
				if len(affectedObjs) == 0 {
					continue
				}
				p.WriteString("The following ")
				if len(affectedObjs) == 1 {
					p.WriteString("resource")
				} else {
					fmt.Fprint(&p, len(affectedObjs))
					p.WriteString(" resources")
				}
				p.WriteString(" are in group '")
				p.WriteString(groupWR.Group)
				p.WriteString("' and will be annotated with the '")
				p.WriteString(groupWR.ReconcileAnnotationKey)
				p.WriteString(": ")
				p.WriteString(groupWR.ReconcileAnnotationValue)
				p.WriteString("' annotation:\n")
				for _, obj := range affectedObjs {
					p.WriteString("\t")
					p.WriteString(utils.ResourceIdentifier(obj))
					p.WriteString("\n")
				}
			}
			p.WriteString("Do you want to continue?")
			if !utils.PromptForConfirmation(p.String(), false) {
				fmt.Println("Command aborted.")
				os.Exit(0)
			}
		}

		errs = []error{}
		for _, key := range sets.List(knownGroups) {
			groupWR := KnownGroupsWithReconcileAnnotation[key]
			affectedObjs := affectedResourcesPerGroup[groupWR]
			for _, obj := range affectedObjs {
				anns := obj.GetAnnotations()
				if anns != nil && anns[groupWR.ReconcileAnnotationKey] == groupWR.ReconcileAnnotationValue {
					if !quiet {
						fmt.Printf("skipping %s because annotation '%s: %s' already exists\n", utils.ResourceIdentifier(obj), groupWR.ReconcileAnnotationKey, groupWR.ReconcileAnnotationValue)
					}
					continue
				}
				if !quiet {
					fmt.Printf("annotated %s with '%s: %s'\n", utils.ResourceIdentifier(obj), groupWR.ReconcileAnnotationKey, groupWR.ReconcileAnnotationValue)
				}
				errs = append(errs, k.Patch(cmd.Context(), obj, client.RawPatch(types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"annotations":{"%s":"%s"}}}`, groupWR.ReconcileAnnotationKey, groupWR.ReconcileAnnotationValue)))))
			}
		}
		if err := errors.Join(errs...); err != nil {
			utils.Fatal(1, "not all reconcile annotations could be applied:\n%s", err.Error())
		}
	},
}

func init() {
	utils.AddDefaultK8sInteractionFlags(ReconcileCmd.PersistentFlags(), k8sOptions)
	ReconcileCmd.Flags().BoolVar(&allResources, "all", false, "Affect all resources of the chosen type(s).")
	ReconcileCmd.Flags().BoolVar(&requireConfirmation, "confirm", false, "If true, command prompts for confirmation before execution.")
	ReconcileCmd.Flags().BoolVarP(&suppressWarnings, "suppress-warnings", "s", false, "If true, no warnings will be printed to stderr if objects of a specific kind could not be listed or resources were not found.")
	ReconcileCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Don't print 'annotated resource ...' messages.")
}

func ValidateReconcileCommand(args []string) {
	if allResources != (len(args) < 2) {
		utils.Fatal(1, "either resource names or --all have to be specified")
	}
}

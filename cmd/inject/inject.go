package inject

import (
	"github.com/spf13/cobra"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
)

// variables for holding the flags
var (
	k8sOptions *utils.K8sInteractionOptions = &utils.K8sInteractionOptions{}
)

// InjectCmd represents the inject command
var InjectCmd = &cobra.Command{
	Use:     "inject",
	Aliases: []string{"in", "i"},
	Short:   "Inject data into k8s resources",
	Long:    `A collection and simplification of common uses of 'kubectl edit'.`,
	GroupID: cmdgroups.ClusterInteraction,
}

func init() {
	InjectCmd.AddCommand(InjectImageCmd)

	utils.AddDefaultK8sInteractionFlags(InjectCmd.PersistentFlags(), k8sOptions)
}

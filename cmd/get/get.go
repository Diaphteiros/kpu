package get

import (
	"github.com/spf13/cobra"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
)

// variables for holding the flags
var (
	k8sOptions *utils.K8sInteractionOptions = &utils.K8sInteractionOptions{}
)

// GetCmd represents the get command
var GetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "Get k8s resources",
	Long:    `A collection of enhanced 'kubectl get' commands.`,
	GroupID: cmdgroups.ClusterInteraction,
}

func init() {
	GetCmd.AddCommand(GetAllCmd)
	GetCmd.AddCommand(GetResourceCmd)
	GetCmd.AddCommand(GetSecretCmd)
	GetCmd.AddCommand(GetPermissionsCmd)

	utils.AddDefaultK8sInteractionFlags(GetCmd.PersistentFlags(), k8sOptions)
	utils.AddK8sImpersonationFlags(GetCmd.PersistentFlags(), k8sOptions)
}

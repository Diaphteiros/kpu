package kubeconfig

import (
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
	"github.com/spf13/cobra"
)

// KubeconfigCmd represents the kubeconfig command
var KubeconfigCmd = &cobra.Command{
	Use:     "kubeconfig",
	Aliases: []string{"kcfg", "k"},
	Short:   "Useful commands for working with kubeconfig files",
	Long:    `Useful commands for working with kubeconfig files.`,
	GroupID: cmdgroups.Kubeconfig,
}

func init() {
	KubeconfigCmd.AddCommand(KubeconfigGetCmd)
}

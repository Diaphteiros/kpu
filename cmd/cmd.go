package cmd

import (
	"github.com/Diaphteiros/kpu/cmd/del"
	"github.com/Diaphteiros/kpu/cmd/finalize"
	"github.com/Diaphteiros/kpu/cmd/get"
	"github.com/Diaphteiros/kpu/cmd/inject"
	"github.com/Diaphteiros/kpu/cmd/kubeconfig"
	"github.com/Diaphteiros/kpu/cmd/reconcile"
	"github.com/Diaphteiros/kpu/cmd/seconds"
	"github.com/Diaphteiros/kpu/cmd/version"
	"github.com/Diaphteiros/kpu/pkg/utils/cmdgroups"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "kpu",
	Short: "Improve your k8s cluster interactions",
	Long: `The 'Kubernetes Power Utilities' aim at supplementing the 'kubectl' CLI
by providing some additional functionality.`,
	DisableAutoGenTag: true,
}

func init() {
	RootCmd.AddGroup(&cobra.Group{ID: cmdgroups.ClusterInteraction, Title: "Cluster Interaction:"})
	RootCmd.AddCommand(del.DeleteCmd)
	RootCmd.AddCommand(finalize.FinalizeCmd)
	RootCmd.AddCommand(get.GetCmd)
	RootCmd.AddCommand(reconcile.ReconcileCmd)
	RootCmd.AddCommand(inject.InjectCmd)
	RootCmd.AddGroup(&cobra.Group{ID: cmdgroups.Kubeconfig, Title: "Kubeconfig:"})
	RootCmd.AddCommand(kubeconfig.KubeconfigCmd)
	RootCmd.AddGroup(&cobra.Group{ID: cmdgroups.Utilities, Title: "Utilities:"})
	RootCmd.AddCommand(seconds.FromSecondsCmd)
	RootCmd.AddCommand(seconds.ToSecondsCmd)
	// generic
	RootCmd.AddCommand(version.VersionCmd)
}

package kubeconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	kcfgPath              string
	namespace             string
	serviceaccount        string
	clusterrolebinding    string
	clusterrole           string
	output                utils.OutputFormat
	minify                bool
	desiredValidityString string
	replace               bool
)

var (
	desiredValidity time.Duration
)

// KubeconfigGetCmd represents the 'kubeconfig get' command
var KubeconfigGetCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Args:    cobra.NoArgs,
	Short:   "Creates a static token kubeconfig for the targeted cluster",
	Long: fmt.Sprintf(`Creates a static token kubeconfig for the targeted cluster.

This requires a ServiceAccount in the cluster for which the token can be created.
Namespace and ServiceAccount will be created if they do not exist.
The ClusterRoleBinding will be created or updated, but if it exists, it needs to have the '%s: %s' label, otherwise the command will fail.
The ClusterRole has to exist already and will not be created or modified.
You can control the names of these resource via the respective flags.

By default, the generated kubeconfig is printed to stdout in %s format.
Use --output (or -o) to change the output format.
If --replace is set, the currently used kubeconfig is replaced instead of printing the result to stdout.

Use --minify to remove all contexts, clusters, and authentication methods that are currently not used from the returned kubeconfig.
Note that this can remove cluster access information from your current kubeconfig when used in combination with the aforementioned --replace flag.
	`, utils.KPUCreatedByLabel, utils.KPUIdentity, utils.OUTPUT_YAML),
	Run: func(cmd *cobra.Command, args []string) {
		ValidateKubeconfigGetCommand(args)

		var err error
		desiredValidity, err = utils.StringToDuration(desiredValidityString)
		if err != nil {
			utils.Fatal(1, "error converting desired validity to duration: %s", err.Error())
		}

		k, err := utils.LoadKubeconfig(kcfgPath)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}
		newKcfg, expTime, err := k.CreateStaticTokenKubeconfig(cmd.Context(), namespace, serviceaccount, clusterrolebinding, clusterrole, int64(desiredValidity.Seconds()), minify)
		if err != nil {
			utils.Fatal(1, "error creating static token kubeconfig: %s", err.Error())
		}

		var kcfgData []byte
		switch output {
		case utils.OUTPUT_YAML:
			kcfgData, err = yaml.Marshal(newKcfg)
			if err != nil {
				utils.Fatal(1, "error converting kubeconfig to yaml: %s", err.Error())
			}
			expirationComment := fmt.Sprintf("# token expires at %s\n", expTime.Format(time.RFC3339))
			kcfgData = append([]byte(expirationComment), kcfgData...)
		case utils.OUTPUT_JSON:
			kcfgData, err = json.Marshal(newKcfg)
			if err != nil {
				utils.Fatal(1, "error converting kubeconfig to json: %s", err.Error())
			}
		}

		if replace {
			os.WriteFile(k.Path, kcfgData, os.ModePerm)
		} else {
			fmt.Println(string(kcfgData))
		}
	},
}

func init() {
	KubeconfigGetCmd.PersistentFlags().StringVar(&kcfgPath, "kubeconfig", "", "Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.")
	KubeconfigGetCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "statictoken", "Namespace for the cluster interaction. Will be created if it does not exist.")
	KubeconfigGetCmd.PersistentFlags().StringVar(&serviceaccount, "serviceaccount", "admin", "ServiceAccount to create the static token. Will be created if it does not exist.")
	KubeconfigGetCmd.PersistentFlags().StringVar(&clusterrolebinding, "clusterrolebinding", "kpu.statictoken", "Name for the ClusterRoleBinding to bind the ServiceAccount to the ClusterRole. Will be overwritten, if it exists and was previously created by this command.")
	KubeconfigGetCmd.PersistentFlags().StringVar(&clusterrole, "clusterrole", "cluster-admin", "Name of the ClusterRole bind the ServiceAccount to. Has to exist and will not be modified. The generated kubeconfig will have the permissions of this ClusterRole.")
	KubeconfigGetCmd.PersistentFlags().BoolVar(&minify, "minify", false, "If true, all contexts, clusters, and authentication methods that are currently not used will be removed from the returned kubeconfig. Be cautious when using this in combination with the --replace flag.")
	KubeconfigGetCmd.PersistentFlags().BoolVarP(&replace, "replace", "r", false, "If true, the currently used kubeconfig is replaced instead of printing the result to stdout.")
	KubeconfigGetCmd.PersistentFlags().StringVarP(&desiredValidityString, "validity", "v", "90d", "Desired validity of the token. Must be a duration string like '1h' or '1y15d3h'. Note that the actual validity might be shorter, depending on the k8s apiserver configuration.")
	utils.AddOutputFlag(KubeconfigGetCmd.PersistentFlags(), &output, utils.OUTPUT_YAML, utils.OUTPUT_JSON, utils.OUTPUT_YAML)
}

func ValidateKubeconfigGetCommand(args []string) {
	if namespace == "" {
		utils.Fatal(1, "namespace must not be empty")
	}
	if serviceaccount == "" {
		utils.Fatal(1, "serviceaccount must not be empty")
	}
	if clusterrolebinding == "" {
		utils.Fatal(1, "clusterrolebinding must not be empty")
	}
	if clusterrole == "" {
		utils.Fatal(1, "clusterrole must not be empty")
	}
	utils.ValidateOutputFormat(output, utils.OUTPUT_JSON, utils.OUTPUT_YAML)
}

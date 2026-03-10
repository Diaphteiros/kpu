package get

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Diaphteiros/kpu/pkg/utils"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/yaml"
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
			resourceNames = []string{}
			for _, arg := range args[1:] {
				resourceNames = append(resourceNames, strings.Split(arg, ",")...)
			}
			slices.Filter(resourceNames[:0], resourceNames, func(s string) bool { return s != "" })
		}

		k, err := utils.LoadKubeconfig(k8sOptions.KubeconfigPath)
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

		switch output {
		case utils.OUTPUT_TEXT:
			t := utils.NewOutputTable[unstructured.Unstructured]()
			if k8sOptions.AllNamespaces {
				t.WithColumn("namespace", func(obj unstructured.Unstructured) string { return obj.GetNamespace() })
			}
			t.WithColumn("name", func(obj unstructured.Unstructured) string { return obj.GetName() })
			t.WithColumn("kind", func(obj unstructured.Unstructured) string { return obj.GetObjectKind().GroupVersionKind().Kind }).
				WithColumn("group", func(obj unstructured.Unstructured) string { return obj.GetObjectKind().GroupVersionKind().Group }).
				WithColumn("version", func(obj unstructured.Unstructured) string { return obj.GetObjectKind().GroupVersionKind().Version })
			t.WithData(objects.Items...)
			fmt.Print(t.String())
		case utils.OUTPUT_JSON:
			data, err := json.MarshalIndent(objects, "", "  ")
			if err != nil {
				utils.Fatal(1, "error converting objects to json: %s", err.Error())
			}
			fmt.Println(string(data))
		case utils.OUTPUT_YAML:
			data, err := yaml.Marshal(objects)
			if err != nil {
				utils.Fatal(1, "error converting objects to yaml: %s", err.Error())
			}
			fmt.Println(string(data))
		}
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

package get

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/yaml"

	"github.com/Diaphteiros/kpu/pkg/utils"
)

// GetSecretCmd represents the 'get secret' command
var GetSecretCmd = &cobra.Command{
	Use:     "secret",
	Aliases: []string{"secrets", "sec", "s"},
	Args:    cobra.MinimumNArgs(0),
	Short:   "Get the specified secrets",
	Long: `Lists/gets secrets, but prints the plain-text 'stringData' field instead of the base64-encoded 'data' one.

Examples:

	> kpu get secret bar -n foo -o yaml
	Returns something like this:
		items:
		- apiVersion: v1
			kind: Secret
			metadata:
				creationTimestamp: "2024-07-12T15:35:30Z"
				name: bar
				namespace: foo
				resourceVersion: "2071574"
				uid: 74a0ef49-91d5-4391-b8f3-659d8fb443e2
			stringData:
				bar: baz
				foo: foobar
			type: Opaque
		metadata: {}
`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateGetSecretCommand(args)

		var resourceNames []string
		if len(args) > 0 {
			resourceNames = []string{} //nolint:prealloc
			for _, arg := range args {
				resourceNames = append(resourceNames, strings.Split(arg, ",")...)
			}
			slices.Filter(resourceNames[:0], resourceNames, func(s string) bool { return s != "" })
		}

		k, err := utils.LoadKubeconfigWithImpersonation(k8sOptions.KubeconfigPath, k8sOptions.ImpersonationConfig)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		objects := &corev1.SecretList{}
		if len(resourceNames) == 1 {
			sec := &corev1.Secret{}
			ns := k8sOptions.Namespace
			if ns == "" {
				ns = k.DefaultNamespace
			}
			if err := k.Get(cmd.Context(), types.NamespacedName{Name: resourceNames[0], Namespace: ns}, sec); err != nil {
				utils.Fatal(1, "error getting secret: %s", err.Error())
			}
			objects.Items = []corev1.Secret{*sec}
		} else {
			if err := k.List(cmd.Context(), objects, k.ConstructListOptions(k8sOptions)...); err != nil {
				utils.Fatal(1, "error listing secrets: %s", err.Error())
			}

			if len(resourceNames) > 0 {
				objects.Items = utils.FilterSlice(objects.Items, func(obj corev1.Secret) bool {
					return slices.Contains(resourceNames, obj.GetName())
				})
			}
		}

		for i := range objects.Items {
			obj := &objects.Items[i]
			if len(obj.Data) > 0 {
				obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))
				obj.StringData = make(map[string]string, len(obj.Data))
				for k, v := range obj.Data {
					obj.StringData[k] = string(v)
				}
				obj.Data = nil
			}
			if !showManagedFields && output != utils.OUTPUT_TEXT {
				// filter out managedFields
				obj.SetManagedFields(nil)
			}
		}

		switch output {
		case utils.OUTPUT_TEXT:
			t := utils.NewOutputTable[corev1.Secret]()
			if k8sOptions.AllNamespaces {
				t.WithColumn("namespace", func(obj corev1.Secret) string { return obj.GetNamespace() })
			}
			t.WithColumn("name", func(obj corev1.Secret) string { return obj.GetName() })
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
	GetSecretCmd.Flags().BoolVar(&showManagedFields, "show-managed-fields", false, "If true, keep the managedFields when printing objects in JSON or YAML format.")
	utils.AddOutputFlag(GetSecretCmd.Flags(), &output, utils.OUTPUT_TEXT)
}

func ValidateGetSecretCommand(args []string) {
	utils.ValidateOutputFormat(output)
}

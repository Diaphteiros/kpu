package get

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	authzv1 "k8s.io/api/authorization/v1"

	"github.com/Diaphteiros/kpu/pkg/utils"
)

// GetPermissionsCmd represents the 'get permissions' command
var GetPermissionsCmd = &cobra.Command{
	Use:     "permissions",
	Aliases: []string{"permission", "perms", "perm", "p"},
	Args:    cobra.NoArgs,
	Short:   "Get the permissions of the current user",
	Long: `Get the permissions of the current user.

This command uses the SelfSubjectRulesReview API to determine the permissions of the current user and prints them.

Examples:

	> kpu get permissions
	Returns all permissions the current user has for the default namespace, formatted as a table.

	> kpu get permissions -A -o json
	Returns all permissions the current user has for all namespaces, formatted as JSON.
`,
	Run: func(cmd *cobra.Command, args []string) {
		k, err := utils.LoadKubeconfig(k8sOptions.KubeconfigPath)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		ssrr := &authzv1.SelfSubjectRulesReview{
			Spec: authzv1.SelfSubjectRulesReviewSpec{},
		}
		if k8sOptions.AllNamespaces {
			ssrr.Spec.Namespace = "*"
		} else if k8sOptions.Namespace != "" {
			ssrr.Spec.Namespace = k8sOptions.Namespace
		} else {
			ssrr.Spec.Namespace = k.DefaultNamespace
		}
		namespace := ssrr.Spec.Namespace // spec gets lost after creation
		ssrr.SetName("kpu-get-permissions")
		if err := k.Create(cmd.Context(), ssrr); err != nil {
			utils.Fatal(1, "error creating SelfSubjectRulesReview: %s", err.Error())
		}

		switch output {
		case utils.OUTPUT_TEXT:
			if ssrr.Status.EvaluationError != "" {
				cmd.Printf("Error evaluating permissions: %s\n", ssrr.Status.EvaluationError)
			}
			cmd.Print("Permissions for resources in ")
			if namespace == "*" {
				cmd.Print("all namespaces")
			} else {
				cmd.Printf("namespace '%s'", namespace)
			}
			if ssrr.Status.Incomplete {
				cmd.Print(" [INCOMPLETE]")
			}
			cmd.Println(":")
			resPerms := dedimensionalizeResourceRules(ssrr.Status.ResourceRules)
			t := utils.NewOutputTable[resourcePermission]()
			t.WithColumn("API Group", func(obj resourcePermission) string { return obj.APIGroup })
			t.WithColumn("Resource", func(obj resourcePermission) string { return obj.Resource })
			t.WithColumn("Verbs", func(obj resourcePermission) string { return strings.Join(obj.Verbs, ", ") })
			t.WithColumn("Resource Names", func(obj resourcePermission) string { return shortenTo(strings.Join(obj.ResourceNames, ", "), 100) })
			t.WithData(resPerms...)
			cmd.Print(t.String())
		case utils.OUTPUT_JSON:
			data, err := json.MarshalIndent(ssrr, "", "  ")
			if err != nil {
				utils.Fatal(1, "error converting SelfSubjectRulesReview to json: %s", err.Error())
			}
			cmd.Println(string(data))
		case utils.OUTPUT_YAML:
			data, err := yaml.Marshal(ssrr)
			if err != nil {
				utils.Fatal(1, "error converting SelfSubjectRulesReview to yaml: %s", err.Error())
			}
			cmd.Println(string(data))
		}
	},
}

func init() {
	utils.AddOutputFlag(GetPermissionsCmd.Flags(), &output, utils.OUTPUT_TEXT)
}

// resourcePermission is a helper struct for 'flattening' the multi-dimensional structure of the ResourceRules struct returned from the SelfSubjectRulesReview API.
type resourcePermission struct {
	APIGroup      string
	Resource      string
	Verbs         []string
	ResourceNames []string
}

func dedimensionalizeResourceRules(rules []authzv1.ResourceRule) []resourcePermission {
	res := []resourcePermission{}
	for _, rule := range rules {
		for _, apiGroup := range rule.APIGroups {
			for _, resource := range rule.Resources {
				tmp := resourcePermission{
					APIGroup: apiGroup,
					Resource: resource,
				}
				if slices.Contains(rule.Verbs, "*") {
					tmp.Verbs = []string{"*"}
				} else {
					tmp.Verbs = rule.Verbs
					slices.Sort(tmp.Verbs)
				}
				if slices.Contains(rule.ResourceNames, "*") {
					tmp.ResourceNames = nil
				} else {
					tmp.ResourceNames = rule.ResourceNames
					slices.Sort(tmp.ResourceNames)
				}
				res = append(res, tmp)
			}
		}
	}

	slices.SortFunc(res, func(a, b resourcePermission) int {
		if a.APIGroup == "*" && b.APIGroup != "*" {
			return -1
		} else if a.APIGroup != "*" && b.APIGroup == "*" {
			return 1
		}
		tmp := strings.Compare(a.APIGroup, b.APIGroup)
		if tmp != 0 {
			return tmp
		}
		if a.Resource == "*" && b.Resource != "*" {
			return -1
		} else if a.Resource != "*" && b.Resource == "*" {
			return 1
		}
		tmp = strings.Compare(a.Resource, b.Resource)
		if tmp != 0 {
			return tmp
		}
		if !slices.Equal(a.Verbs, b.Verbs) {
			if slices.Equal(a.Verbs, []string{"*"}) {
				return -1
			} else if slices.Equal(b.Verbs, []string{"*"}) {
				return 1
			}
			return len(b.Verbs) - len(a.Verbs)
		}
		if len(a.ResourceNames) != 0 && len(b.ResourceNames) != 0 {
			return len(b.ResourceNames) - len(a.ResourceNames)
		}
		if len(a.ResourceNames) == 0 {
			return -1
		} else if len(b.ResourceNames) == 0 {
			return 1
		}
		return 0
	})
	return res
}

func shortenTo(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-1] + "…"
}

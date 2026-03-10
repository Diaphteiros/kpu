package version

import (
	"encoding/json"
	"fmt"

	"github.com/Diaphteiros/kpu/pkg/utils"
	staticversion "github.com/Diaphteiros/kpu/pkg/version"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// variables for holding the flags
var (
	output utils.OutputFormat
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Args:    cobra.NoArgs,
	Short:   "Print the version",
	Long: `Output the version of the CLI.

Examples:

	> kpu version
	v1.2.3
	
	> kpu version -o json
	{"major":"v1","minor":"2","gitVersion":"v1.2.3","gitCommit":"76c01d5337fc9de6e053b4e5bafd5239c8b7a973","gitTreeState":"dirty","buildDate":"2024-04-26T11:29:39+02:00","goVersion":"go1.22.2","compiler":"gc","platform":"darwin/arm64"}

	> kpu version -o yaml
	buildDate: "2024-04-26T11:29:39+02:00"
	compiler: gc
	gitCommit: 76c01d5337fc9de6e053b4e5bafd5239c8b7a973
	gitTreeState: dirty
	gitVersion: v1.2.3
	goVersion: go1.22.2
	major: v1
	minor: "2"
	platform: darwin/arm64`,
	Run: func(cmd *cobra.Command, args []string) {
		switch output {
		case utils.OUTPUT_TEXT:
			fmt.Print(staticversion.Version.String())
		case utils.OUTPUT_JSON:
			data, err := json.Marshal(staticversion.Version)
			if err != nil {
				utils.Fatal(1, "error converting version to json: %s\n", err.Error())
			}
			fmt.Println(string(data))
		case utils.OUTPUT_YAML:
			data, err := yaml.Marshal(staticversion.Version)
			if err != nil {
				utils.Fatal(1, "error converting version to yaml: %s\n", err.Error())
			}
			fmt.Print(string(data))
		default:
			utils.Fatal(1, "unknown output format '%s'", string(output))
		}
	},
}

func init() {
	utils.AddOutputFlag(VersionCmd.Flags(), &output, utils.OUTPUT_TEXT)
}

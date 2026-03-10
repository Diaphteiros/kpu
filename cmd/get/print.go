package get

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"

	"github.com/Diaphteiros/kpu/pkg/utils"
)

func printUnstructuredObjects(output utils.OutputFormat, objects *unstructured.UnstructuredList) {
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
}

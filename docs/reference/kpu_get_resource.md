## kpu get resource

Get the specified k8s resources

### Synopsis

This is pretty similar to 'kubectl get' and mainly in here for debugging purposes.

Examples:

	> kpu get resource namespace foo -o yaml
	Returns the namespace 'foo' in yaml format.

	> kpu get resource pod,deployment -n foo -l example.org/mylabel
	Lists all pods and deployments in the namespace 'foo' with the label 'example.org/mylabel' (independent of the label value).
	

```
kpu get resource [flags]
```

### Options

```
  -h, --help                  help for resource
  -o, --output string         Output format. Valid formats are [json, text, yaml]. (default "text")
      --show-managed-fields   If true, keep the managedFields when printing objects in JSON or YAML format.
```

### Options inherited from parent commands

```
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
```

### SEE ALSO

* [kpu get](kpu_get.md)	 - Get k8s resources


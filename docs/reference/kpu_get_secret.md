## kpu get secret

Get the specified secrets

### Synopsis

Lists/gets secrets, but prints the plain-text 'stringData' field instead of the base64-encoded 'data' one.

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


```
kpu get secret [flags]
```

### Options

```
  -h, --help                  help for secret
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


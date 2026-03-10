## kpu get

Get k8s resources

### Synopsis

A collection of enhanced 'kubectl get' commands.

### Options

```
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
  -h, --help                help for get
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions
* [kpu get all](kpu_get_all.md)	 - Get all k8s resources
* [kpu get resource](kpu_get_resource.md)	 - Get the specified k8s resources
* [kpu get secret](kpu_get_secret.md)	 - Get the specified secrets


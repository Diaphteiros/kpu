## kpu inject

Inject data into k8s resources

### Synopsis

A collection and simplification of common uses of 'kubectl edit'.

### Options

```
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
  -h, --help                help for inject
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions
* [kpu inject image](kpu_inject_image.md)	 - Replace the image of a container in a pod, deployment, statefulset, or daemonset, or in one of the OpenMCP resources platformservice, serviceprovider, or clusterprovider


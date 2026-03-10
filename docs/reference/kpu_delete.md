## kpu delete

Delete k8s resources

### Synopsis

This command works basically like 'kubectl delete', with two major differences:
1. It lists all affected resources and asks for confirmation beforehand (unless -y is specified).
2. It adds deletion confirmation annotations for specific API groups before attempting the deletion.

Currently, the following deletion confirmation rules are implemented:
- core.gardener.cloud => confirmation.gardener.cloud/deletion=true
- openmcp.cloud       => confirmation.openmcp.cloud/deletion=true


```
kpu delete type1[,type2,...] [name1,name2 [name3 ...]] [flags]
```

### Options

```
      --all                 Affect all resources of the chosen type(s).
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
  -h, --help                help for delete
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
  -s, --suppress-warnings   If true, no warnings will be printed to stderr if objects of a specific kind could not be listed or resources were not found.
  -y, --yes                 If true, command won't prompt for confirmation.
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions


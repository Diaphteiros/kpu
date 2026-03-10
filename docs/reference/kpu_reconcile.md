## kpu reconcile

Add reconcile annotations to k8s resources

### Synopsis

This command patches reconcile annotations to the specified resources.

Known api groups are:
	core.gardener.cloud       => gardener.cloud/operation: reconcile
	landscaper.gardener.cloud => landscaper.gardener.cloud/operation: reconcile
	openmcp.cloud             => openmcp.cloud/operation: reconcile
	core.openmcp.cloud        => openmcp.cloud/operation: reconcile

Examples:

	> kpu reconcile installation foo -n bar
	Adds the 'landscaper.gardener.cloud/operation: reconcile' annotation to the installation 'foo' in the namespace 'bar'.
	
	> kpu reconcile shoot --all -n garden-myproject
	Adds the 'gardener.cloud/operation: reconcile' annotation to all shoots in the namespace 'garden-myproject'.
	

```
kpu reconcile [type1[,type2,...]] [name1,name2 [name3 ...]] [flags]
```

### Options

```
      --all                 Affect all resources of the chosen type(s).
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
      --confirm             If true, command prompts for confirmation before execution.
  -h, --help                help for reconcile
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -q, --quiet               Don't print 'annotated resource ...' messages.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
  -s, --suppress-warnings   If true, no warnings will be printed to stderr if objects of a specific kind could not be listed or resources were not found.
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions


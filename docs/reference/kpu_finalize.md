## kpu finalize

Remove finalizers from k8s resources

### Synopsis

This command can remove all or only specific finalizers from specific or all resources.

Examples:

	> kpu finalize secret foo -n bar
	Removes all finalizers from the secret 'foo' in the namespace 'bar'.
	
	> kpu finalize --all -n bar
	Removes all finalizers from all resources in the namespace 'bar'.
	
	> kpu finalize -f foo.bar.baz/foobar --all -A
	Removes the finalizer 'foo.bar.baz/foobar' from all resources in all namespaces.
	Note that this affects only namespace-scoped resources due to the default value 'namespaced' for --scope.

	> kpu finalize secret,configmap foo bar -f foo.bar.baz/foobar -f foo.bar.baz/asdf
	Removes the finalizers 'foo.bar.baz/foobar' and 'foo.bar.baz/asdf' from the secrets and configmaps
	named 'foo' and 'bar' in the default namespace.
	

```
kpu finalize [type1[,type2,...]] [name1,name2 [name3 ...]] [flags]
```

### Options

```
      --all                 Affect all resources of the chosen type(s).
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
  -f, --finalizer strings   Name(s) of finalizer(s) to be removed. Leave empty to remove all finalizers.
  -h, --help                help for finalize
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
      --scope string        Resource scope for the command. Valid scopes are [all, cluster, namespaced]. (default "namespaced")
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
  -s, --suppress-warnings   If true, no warnings will be printed to stderr if objects of a specific kind could not be listed or resources were not found.
  -y, --yes                 If true, command won't prompt for confirmation.
```

### SEE ALSO

* [kpu](kpu.md)	 - Improve your k8s cluster interactions


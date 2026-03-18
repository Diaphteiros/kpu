## kpu get permissions

Get the permissions of the current user

### Synopsis

Get the permissions of the current user.

This command uses the SelfSubjectRulesReview API to determine the permissions of the current user and prints them.

Examples:

	> kpu get permissions
	Returns all permissions the current user has for the default namespace, formatted as a table.

	> kpu get permissions -A -o json
	Returns all permissions the current user has for all namespaces, formatted as JSON.

	> kpu get permissions -n foo --as my-user
	Returns all permissions the user 'my-user' has for the namespace 'foo', formatted as a table.


```
kpu get permissions [flags]
```

### Options

```
  -h, --help            help for permissions
  -o, --output string   Output format. Valid formats are [json, text, yaml]. (default "text")
```

### Options inherited from parent commands

```
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
      --as string           Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group strings    Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string       UID to impersonate for the operation.
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
```

### SEE ALSO

* [kpu get](kpu_get.md)	 - Get k8s resources


## kpu kubeconfig get

Creates a static token kubeconfig for the targeted cluster

### Synopsis

Creates a static token kubeconfig for the targeted cluster.

This requires a ServiceAccount in the cluster for which the token can be created.
Namespace and ServiceAccount will be created if they do not exist.
The ClusterRoleBinding will be created or updated, but if it exists, it needs to have the 'created-by: kpu' label, otherwise the command will fail.
The ClusterRole has to exist already and will not be created or modified.
You can control the names of these resource via the respective flags.

By default, the generated kubeconfig is printed to stdout in yaml format.
Use --output (or -o) to change the output format.
If --replace is set, the currently used kubeconfig is replaced instead of printing the result to stdout.

Use --minify to remove all contexts, clusters, and authentication methods that are currently not used from the returned kubeconfig.
Note that this can remove cluster access information from your current kubeconfig when used in combination with the aforementioned --replace flag.
	

```
kpu kubeconfig get [flags]
```

### Options

```
      --clusterrole string          Name of the ClusterRole bind the ServiceAccount to. Has to exist and will not be modified. The generated kubeconfig will have the permissions of this ClusterRole. (default "cluster-admin")
      --clusterrolebinding string   Name for the ClusterRoleBinding to bind the ServiceAccount to the ClusterRole. Will be overwritten, if it exists and was previously created by this command. (default "kpu.statictoken")
  -h, --help                        help for get
      --kubeconfig string           Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
      --minify                      If true, all contexts, clusters, and authentication methods that are currently not used will be removed from the returned kubeconfig. Be cautious when using this in combination with the --replace flag.
  -n, --namespace string            Namespace for the cluster interaction. Will be created if it does not exist. (default "statictoken")
  -o, --output string               Output format. Valid formats are [json, yaml]. (default "yaml")
  -r, --replace                     If true, the currently used kubeconfig is replaced instead of printing the result to stdout.
      --serviceaccount string       ServiceAccount to create the static token. Will be created if it does not exist. (default "admin")
  -v, --validity string             Desired validity of the token. Must be a duration string like '1h' or '1y15d3h'. Note that the actual validity might be shorter, depending on the k8s apiserver configuration. (default "90d")
```

### SEE ALSO

* [kpu kubeconfig](kpu_kubeconfig.md)	 - Useful commands for working with kubeconfig files


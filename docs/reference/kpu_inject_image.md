## kpu inject image

Replace the image of a container in a pod, deployment, statefulset, or daemonset, or in one of the OpenMCP resources platformservice, serviceprovider, or clusterprovider

### Synopsis

Replace the image of a container in a pod, deployment, statefulset, or daemonset, or in one of the OpenMCP resources platformservice, serviceprovider, or clusterprovider.

By default, the command targets deployments. Use the --resource-type flag to change this.
You can specify either a complete image, consisting of the repository followed by either ':' and a tag or '@' and a digest,
or just specify the tag or digest. In the latter case, the repository of the existing image will be used.
If you prefix the tag/digest with ':'/'@', the respective separator will be used, otherwise the one from the existing image will be used.

This means that these are valid formats for `<image-replacement>`:
- `<image-repo>:<image-tag>`
- `<image-repo>@<image-digest>`
- `:<image-tag>`
- `@<image-digest>`
- `<image-tag>`
- `<image-digest>`

The command prompts you to choose the container to replace the image in.

Use '--init' to chose from init containers instead of regular containers.

Examples:

	> kpu inject image mydeploy myimage:latest
	Replace the image in a container of the deployment 'mydeploy' with 'myimage:latest'.

	> kpu inject image -n foo mydeploy :latest
	Replace the currently used image tag/digest in a container of the deployment 'mydeploy' in namespace 'foo' with the 'latest' tag.

	> kpu inject image mydeploy latest
	Replace the image in a container of the deployment 'mydeploy' with the 'latest' tag.
	If the current image uses a digest and not a tag, the resulting image will be '<image-repo>@latest', which is probably not what you want.

	> kpu inject image mypod :latest --resource-type pod
	Replace the currently used image tag/digest in a container of the pod 'mypod' with the 'latest' tag.
	

```
kpu inject image <resource-name> `<image-replacement>` [flags]
```

### Options

```
  -h, --help                   help for image
      --init                   Inject the image into an init container instead of a regular container.
  -t, --resource-type string   Which resource to inject the image into. Valid values are [pod, deployment, statefulset, daemonset, platformservice, serviceprovider, clusterprovider]. (default "deployment")
```

### Options inherited from parent commands

```
  -A, --all-namespaces      If true, the command affects all namespaces. Overwrites --namespace flag.
      --kubeconfig string   Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.
  -n, --namespace string    Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.
  -l, --selector strings    Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.
```

### SEE ALSO

* [kpu inject](kpu_inject.md)	 - Inject data into k8s resources


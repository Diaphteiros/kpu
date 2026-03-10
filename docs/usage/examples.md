# Usage Examples

This is just a small overview over some of the available commands. For the full list, check out the generated [command reference](docs/reference/kpu.md).

## Subcommands

### kpu get all

Have you ever tried to run `kubectl get all`, were confused that it does actually not list _all_ resources (e.g. no secrets), have then googled and found out that all github issues regarding this topic were closed with 'we won't change that due to backward compatibility' despite users continuing to be confused by this behavior?

If so, then `kpu get all` does what `kubectl get all` was supposed to do: It lists _all_ resources. You can optionally specify whether if you want cluster-scoped or namespace-scoped resources only.

**Example:**
```
> kpu get all -n test
NAME               KIND                  GROUP            VERSION
kube-root-ca.crt   ConfigMap                              v1
test-access        Secret                                 v1
default            ServiceAccount                         v1
```

### kpu get secret

The `kpu get secret` subcommand fetches a secret but converts the base64-encoded `data` part into the clear-text `stringData` field.

**Example:**
```
> kpu get secret test-access -n test -o yaml
items:
- apiVersion: v1
  kind: Secret
  metadata:
    creationTimestamp: "2024-05-16T07:29:27Z"
    name: test-access
    namespace: test
    resourceVersion: "22941721"
    uid: 5d6edd70-d184-4c3f-9817-04110d6d565c
  stringData:
    config: |
      apiVersion: v1
      clusters:
      - cluster:
      ...
```

### kpu finalize

Probably the most interesting subcommand, `kpu finalize` is a powerful tool for removing finalizers from resources in kubernetes. It can remove all or only specific finalizers from one, multiple, or all resources of a specific kind, across namespaces.

Note that, unless deactivated with the `-y` flag, the command prompts the user for confirmation before doing anything.

**Example:**
```
> kpu finalize secret,auth,authz,dp -n test --all -f dependency.example.org/finalizer
This will remove all dependency.example.org/finalizer finalizers from the following 3 resources:
Secret.v1	test/test-access
MyResource.v1alpha1.example.org	test/test
MyOtherResource.v1alpha1.example.org	test/test
Do you want to continue? Confirm with 'y' or 'yes' (case-insensitive):
```

### kpu delete

`kpu delete` is similar to `kubectl delete`, with two major advantages:
- Unless configured otherwise, it lists all affected resources and prompts for confirmation before deleting anything.
- For some resources, which require to be annotated before deletion, the command includes the annotation logic.
  - This is hard-coded and only implemented for specific api groups.

**Example:**
```
> kpu delete quota --all -A
This will delete the following 3 resources:
ResourceQuota.v1	cumulative/cumulative-quota
ResourceQuota.v1	maximum/maximum-quota
ResourceQuota.v1	singular/singular-quota
Do you want to continue? Confirm with 'y' or 'yes' (case-insensitive):
```

### kpu reconcile

This subcommand can patch a reconcile annotation onto all given resources.
The annotation is inferred from the resource's api groups. Check the help for a list of supported api groups.

Note that this command does not prompt for confirmation by default, but can be instructed to do so by using the `--confirm` flag.

**Example:**
```
> kpu reconcile shoot test -n test
annotated Shoot.v1beta1.core.gardener.cloud test/test with 'gardener.cloud/operation: reconcile'
```

## Common Options

Most of the subcommands support a similar set of flags, some of which are mentioned here.

### --kubeconfig

Similar to `kubectl`, a path to the kubeconfig can be provided via the `--kubeconfig` flag. If not set, the `KUBECONFIG` environment variable is read and if that is not set either, the default kubeconfig location (`$HOME/.kube/config`) is used.

### -n / --namespace

Also similar to `kubectl`, use `-n`/`--namespace` to set the namespace for the command. If not set, the default from the kubeconfig is used, or `default` if none is specified there.

### -A / --all-namespaces

Again similar to `kubectl`, `-A`/`--all-namespaces` can be used to make the command affect resources in multiple namespaces.

### --all

Some commands support selecting all resources of the given kind(s) instead of having to name them by specifying the `--all` flag. This behavior is modeled after the corresponding flag from `kubectl delete`.

### -l / --selector

Similar to many `kubectl` commands, one or more label selectors can be used to filter the affected resources by specifying the `-l`/`--selector` flag.

The following modes are supported:
- `-l <key>=<value>` or `-l <key>==<value>` returns only resources with the `<key>: <value>` label
- `-l <key>!=<value>` returns only resources where the `<key>` label does not have the value `<value>`
- `-l <key>` returns only resources with the `<key>` label, independent of its value
- `-l !<key>` returns only resources without the `<key>` label, independent of its value
  - You might have to quote the exclamation mark like this `-l '!<key>'`, depending on your shell.

package utils

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	openmcpproviderv1alpha1 "github.com/openmcp-project/openmcp-operator/api/provider/v1alpha1"
	"github.com/spf13/pflag"
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	k8sapi "k8s.io/client-go/tools/clientcmd/api"
	k8sapiv1 "k8s.io/client-go/tools/clientcmd/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Diaphteiros/kpu/pkg/fork"
)

const (
	rawLabelKeyRegex        = `(?P<key>(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])|([a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]))`
	rawLabelMultiValueRegex = `(?P<values>((([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?)(;(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?)*)`
	rawLabelOperatorRegex   = `(?P<operator>\=\=|\!\=|\=)`

	KPUCreatedByLabel = "created-by"
	KPUIdentity       = "kpu"
)

var (
	// LabelKeyExistsRegex matches a label selector in the form of <key> (label exists) or !<key> (label does not exist).
	LabelKeyExistsRegex = regexp.MustCompile(fmt.Sprintf(`^(?P<mod>\!?)%s$`, rawLabelKeyRegex))
	// LabelKeyOperatorValueRegex matches a label selector in the form of <key><operator><value>.
	LabelKeyOperatorValuesRegex = regexp.MustCompile(fmt.Sprintf(`^%s%s%s$`, rawLabelKeyRegex, rawLabelOperatorRegex, rawLabelMultiValueRegex))

	// indices for the capture groups in the LabelKeyExistsRegex
	LabelKeyExistsRegex_ModIndex = -1
	LabelKeyExistsRegex_KeyIndex = -1

	// indices for the capture groups in the LabelKeyOperatorValueRegex
	LabelKeyOperatorValueRegex_KeyIndex      = -1
	LabelKeyOperatorValueRegex_OperatorIndex = -1
	LabelKeyOperatorValueRegex_ValueIndex    = -1
)

func init() {
	for i, name := range LabelKeyExistsRegex.SubexpNames() {
		switch name {
		case "mod":
			LabelKeyExistsRegex_ModIndex = i
		case "key":
			LabelKeyExistsRegex_KeyIndex = i
		}
	}
	if LabelKeyExistsRegex_ModIndex == -1 || LabelKeyExistsRegex_KeyIndex == -1 {
		panic("could not find capture groups in LabelKeyExistsRegex")
	}

	for i, name := range LabelKeyOperatorValuesRegex.SubexpNames() {
		switch name {
		case "key":
			LabelKeyOperatorValueRegex_KeyIndex = i
		case "operator":
			LabelKeyOperatorValueRegex_OperatorIndex = i
		case "values":
			LabelKeyOperatorValueRegex_ValueIndex = i
		}
	}
	if LabelKeyOperatorValueRegex_KeyIndex == -1 || LabelKeyOperatorValueRegex_OperatorIndex == -1 || LabelKeyOperatorValueRegex_ValueIndex == -1 {
		panic("could not find capture groups in LabelKeyOperatorValuesRegex")
	}
}

type K8sInteractionOptions struct {
	KubeconfigPath      string
	Namespace           string
	AllNamespaces       bool
	LabelSelector       []string
	ImpersonationConfig rest.ImpersonationConfig
}

// Kubeconfig stores information loaded from the kubeconfig.
// It is a wrapper around client.Client and can be used to communicate with the cluster.
type Kubeconfig struct {
	client.Client
	DiscoveryClient   *discovery.DiscoveryClient
	ShortNameExpander meta.RESTMapper
	DefaultNamespace  string
	RESTConfig        *rest.Config
	Path              string
	Raw               k8sapi.Config
}

// AddDefaultK8sInteractionFlags adds default flags for k8s cluster interaction to the given flagset.
// The flags bind to variables in the given K8sInteractionOptions struct.
func AddDefaultK8sInteractionFlags(fs *pflag.FlagSet, ko *K8sInteractionOptions) {
	fs.StringVar(&ko.KubeconfigPath, "kubeconfig", "", "Path to the kubeconfig file. Defaults to $KUBECONFIG or ~/.kube/config, if not set.")
	fs.StringVarP(&ko.Namespace, "namespace", "n", "", "Namespace for the cluster interaction. Defaults to namespace from kubeconfig or 'default', if not set.")
	fs.BoolVarP(&ko.AllNamespaces, "all-namespaces", "A", false, "If true, the command affects all namespaces. Overwrites --namespace flag.")
	fs.StringSliceVarP(&ko.LabelSelector, "selector", "l", nil, "Selector (label query) to filter on. Supports 'exists' (<label>), 'does not exist' (!<label>), 'in' (<label>=<value1>;<value2>) and 'not in' (<label>!=<value1>;<value2>) operators. Allows multiple selectors (comma-separated or setting flag multiple times), they will be ANDed.")
}

// AddK8sImpersonationFlags adds default flags for impersonation to the given flagset.
// The flags bind the ImpersonationConfig in the given K8sInteractionOptions struct.
func AddK8sImpersonationFlags(fs *pflag.FlagSet, ko *K8sInteractionOptions) {
	fs.StringVar(&ko.ImpersonationConfig.UserName, "as", "", "Username to impersonate for the operation. User could be a regular user or a service account in a namespace.")
	fs.StringSliceVar(&ko.ImpersonationConfig.Groups, "as-group", nil, "Group to impersonate for the operation, this flag can be repeated to specify multiple groups.")
	fs.StringVar(&ko.ImpersonationConfig.UID, "as-uid", "", "UID to impersonate for the operation.")
}

// LoadKubeconfig loads a kubeconfig from the given path.
// If the path is empty, it is defaulted to the value of the KUBECONFIG env var first,
// and then to $HOME/.kube/config, if it is still empty.
func LoadKubeconfig(kcfgPath string) (*Kubeconfig, error) {
	return LoadKubeconfigWithImpersonation(kcfgPath, rest.ImpersonationConfig{})
}

func LoadKubeconfigWithImpersonation(kcfgPath string, impersonationConfig rest.ImpersonationConfig) (*Kubeconfig, error) {
	if kcfgPath == "" {
		// 1st fallback: KUBECONFIG env var
		kcfgPath = os.Getenv("KUBECONFIG")
	}
	if kcfgPath == "" {
		// 2nd fallback: $HOME/.kube/config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error determining user home directory: %w", err)
		}
		kcfgPath = path.Join(homeDir, ".kube", "config")
	}
	data, err := os.ReadFile(kcfgPath)
	if err != nil {
		return nil, fmt.Errorf("error reading kubeconfig file: %w", err)
	}
	kcfg, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("error constructing kubeconfig: %w", err)
	}
	res := &Kubeconfig{
		Path: kcfgPath,
	}
	res.RESTConfig, err = kcfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error constructing rest client: %w", err)
	}
	res.RESTConfig.Impersonate = impersonationConfig
	res.Raw, err = kcfg.MergedRawConfig()
	if err != nil {
		return nil, fmt.Errorf("error constructing merged raw kubeconfig: %w", err)
	}
	res.DefaultNamespace = res.Raw.Contexts[res.Raw.CurrentContext].Namespace
	if res.DefaultNamespace == "" {
		res.DefaultNamespace = "default"
	}
	sc := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(sc); err != nil {
		return nil, fmt.Errorf("error adding client-go to scheme: %w", err)
	}
	if err := openmcpproviderv1alpha1.AddToScheme(sc); err != nil {
		return nil, fmt.Errorf("error adding openmcp provider v1alpha1 to scheme: %w", err)
	}
	res.Client, err = client.New(res.RESTConfig, client.Options{Scheme: sc})
	if err != nil {
		return nil, fmt.Errorf("error constructing client: %w", err)
	}
	return res, nil
}

// DiscoverAPIResources takes a rest config and returns all discovered api resources.
// The resources are sorted by group, then by kind, then by version.
func (k *Kubeconfig) DiscoverAPIResources() ([]metav1.APIResource, error) {
	if k.DiscoveryClient == nil {
		dc, err := discovery.NewDiscoveryClientForConfig(k.RESTConfig)
		if err != nil {
			return nil, fmt.Errorf("error constructing discovery client: %w", err)
		}
		k.DiscoveryClient = dc
	}
	_, apiResourceGroups, err := k.DiscoveryClient.ServerGroupsAndResources()
	if err != nil {
		return nil, fmt.Errorf("error discovering groups and resources: %w", err)
	}

	res := []metav1.APIResource{}
	for _, group := range apiResourceGroups {
		fallback, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			return nil, fmt.Errorf("error parsing GroupVersion '%s': %w", group.GroupVersion, err)
		}
		for _, resource := range group.APIResources {
			if resource.Group == "" {
				resource.Group = fallback.Group
			}
			if resource.Version == "" {
				resource.Version = fallback.Version
			}
			res = append(res, resource)
		}
	}

	// remove duplicates
	res = slices.CompactFunc(res, func(a, b metav1.APIResource) bool {
		return a.Kind == b.Kind && a.Group == b.Group && a.Version == b.Version
	})

	// sort results
	slices.SortFunc(res, func(a, b metav1.APIResource) int {
		cmp := strings.Compare(a.Group, b.Group)
		if cmp == 0 {
			cmp = strings.Compare(a.Kind, b.Kind)
		}
		if cmp == 0 {
			cmp = strings.Compare(a.Version, b.Version)
		}
		return cmp
	})

	return res, nil
}

// GetNamespaces lists all k8s namespaces.
func (k *Kubeconfig) GetNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	nsl := &corev1.NamespaceList{}
	if err := k.List(ctx, nsl); err != nil {
		return nil, err
	}
	return nsl.Items, nil
}

// GetAllResources fetches all resources.
// Scope determines if only namespace-scoped, cluster-scoped, or all resources are fetched.
// k8sOptions are used to limit the search to specific namespaces or similar.
func (k *Kubeconfig) GetAllResources(ctx context.Context, scope Scope, k8sOptions *K8sInteractionOptions) ([]unstructured.Unstructured, []error) {
	apis, err := k.DiscoverAPIResources()
	if err != nil {
		Fatal(1, "%s", err.Error())
	}
	errs := []error{}
	objects := []unstructured.Unstructured{}
	for _, api := range apis {
		if api.Namespaced && !scope.IncludesNamespaced() {
			continue
		}
		if !api.Namespaced && !scope.IncludesCluster() {
			continue
		}
		if !slices.Contains(api.Verbs, "list") {
			// don't try to list resources which cannot be listed
			continue
		}
		objL := &unstructured.UnstructuredList{}
		objL.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   api.Group,
			Version: api.Version,
			Kind:    api.Kind,
		})
		if err := k.List(ctx, objL, k.ConstructListOptions(k8sOptions)...); err != nil {
			errs = append(errs, fmt.Errorf("unable to list objects for kind '%s' / group '%s' / version '%s': %w", api.Kind, api.Group, api.Version, err))
		}
		objects = append(objects, objL.Items...)
	}
	return objects, errs
}

// GetGVKForResource resolves a shortname to a GroupVersionKind.
func (k *Kubeconfig) GetGVKForResource(resource string) (schema.GroupVersionKind, error) {
	if k.DiscoveryClient == nil {
		dc, err := discovery.NewDiscoveryClientForConfig(k.RESTConfig)
		if err != nil {
			return schema.GroupVersionKind{}, fmt.Errorf("error constructing discovery client: %w", err)
		}
		k.DiscoveryClient = dc
	}
	if k.ShortNameExpander == nil {
		k.ShortNameExpander = fork.NewShortcutExpander(k.RESTMapper(), k.DiscoveryClient, func(s string) {})
	}
	gvr, err := k.ShortNameExpander.KindFor(schema.GroupVersionResource{Resource: resource})
	if err != nil {
		return schema.GroupVersionKind{}, fmt.Errorf("error inferring GVK from resource name: %w", err)
	}
	return gvr, nil
}

// ListResources lists all resources matching the specified criteria.
// If resourceTypes is empty, all resources are listed (filtered by scope). Otherwise, only the specified types are listed.
// If resourceNames is empty, all resources are listed. Otherwise, only the resources with names contained in the specified list are listed.
func (k *Kubeconfig) ListResources(ctx context.Context, resourceTypes []string, resourceNames []string, scope Scope, k8sOptions *K8sInteractionOptions) ([]unstructured.Unstructured, []error) {
	objects := map[schema.GroupVersionKind][]unstructured.Unstructured{}
	var errs []error
	if len(resourceTypes) == 0 {
		objects[schema.GroupVersionKind{}], errs = k.GetAllResources(ctx, scope, k8sOptions)
	} else {
		errs = []error{}
		for _, rt := range resourceTypes {
			gvk, err := k.GetGVKForResource(rt)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			objL := &unstructured.UnstructuredList{}
			objL.SetGroupVersionKind(gvk)
			if err := k.List(ctx, objL, k.ConstructListOptions(k8sOptions)...); err != nil {
				errs = append(errs, err)
				continue
			}
			if len(objL.Items) > 0 {
				if objects[gvk] == nil {
					objects[gvk] = []unstructured.Unstructured{}
				}
				objects[gvk] = append(objects[gvk], objL.Items...)
			}
		}
	}

	// TODO: better filtering
	if len(resourceNames) > 0 {
		nameSet := sets.New[string](resourceNames...)
		for gvk, objs := range objects {
			// filter out resources which don't have one of the specified names
			objs = slices.DeleteFunc(objs, func(obj unstructured.Unstructured) bool { return !nameSet.Has(obj.GetName()) })
			objects[gvk] = objs
			// check if all specified names were found
			foundNames := sets.New[string]()
			for _, obj := range objs {
				foundNames.Insert(obj.GetName())
			}
			missingNames := nameSet.Difference(foundNames)
			if len(missingNames) > 0 {
				for _, name := range sets.List(missingNames) {
					errs = append(errs, apierrors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, name)) // technically wrong, because resource != kind, but since it is just for error reporting, doesn't really matter here
				}
			}
		}
	}

	finalObjs := []unstructured.Unstructured{} //nolint:prealloc
	for _, objs := range objects {
		finalObjs = append(finalObjs, objs...)
	}

	return finalObjs, errs
}

func (k *Kubeconfig) ConstructListOptions(opts *K8sInteractionOptions) []client.ListOption {
	res := []client.ListOption{}
	if len(opts.LabelSelector) > 0 {
		ls, err := ParseLabelSelectors(opts.LabelSelector)
		if err != nil {
			Fatal(1, "%s", err.Error())
		}
		res = append(res, client.MatchingLabelsSelector{Selector: ls})
	}
	if opts.AllNamespaces {
		return res
	}
	ns := opts.Namespace
	if ns == "" {
		ns = k.DefaultNamespace
	}
	res = append(res, client.InNamespace(ns))
	return res
}

// ParseLabelSelectors parses a list of label selectors into a labels.Selector.
// Each individual selector must be in the form of <key><operator><value> or [!]<key>.
// Possible values for <operator> are ==, =, !=.
func ParseLabelSelectors(selectors []string) (labels.Selector, error) {
	res := &metav1.LabelSelector{}
	for _, sel := range selectors {
		// check for label selectors in the form of <key><operator><values>
		matches := LabelKeyOperatorValuesRegex.FindAllStringSubmatch(sel, -1)
		if len(matches) > 0 {
			me := metav1.LabelSelectorRequirement{
				Key:    matches[0][LabelKeyOperatorValueRegex_KeyIndex],
				Values: strings.Split(matches[0][LabelKeyOperatorValueRegex_ValueIndex], ";"),
			}
			switch matches[0][LabelKeyOperatorValueRegex_OperatorIndex] {
			case "==":
				fallthrough
			case "=":
				me.Operator = metav1.LabelSelectorOpIn
			case "!=":
				me.Operator = metav1.LabelSelectorOpNotIn
			default:
				return nil, fmt.Errorf("invalid operator: %s", matches[0][LabelKeyOperatorValueRegex_OperatorIndex])
			}
			if res.MatchExpressions == nil {
				res.MatchExpressions = []metav1.LabelSelectorRequirement{}
			}
			res.MatchExpressions = append(res.MatchExpressions, me)
			continue
		}
		// check for label selectors in the form of [!]<key>
		matches = LabelKeyExistsRegex.FindAllStringSubmatch(sel, -1)
		if len(matches) > 0 {
			me := metav1.LabelSelectorRequirement{
				Key: matches[0][LabelKeyExistsRegex_KeyIndex],
			}
			if matches[0][LabelKeyExistsRegex_ModIndex] == "!" {
				me.Operator = metav1.LabelSelectorOpDoesNotExist
			} else {
				me.Operator = metav1.LabelSelectorOpExists
			}
			if res.MatchExpressions == nil {
				res.MatchExpressions = []metav1.LabelSelectorRequirement{}
			}
			res.MatchExpressions = append(res.MatchExpressions, me)
			continue
		}

		return nil, fmt.Errorf("invalid label selector: %s", sel)
	}

	return metav1.LabelSelectorAsSelector(res)
}

// CreateStaticTokenKubeconfig creates a kubeconfig that uses a serviceaccount token for authentication.
// For this, it performs the following steps:
// - create the namespace if it does not exist
// - create the serviceaccount if it does not exist
// - create/update the clusterrolebinding (if it exists and was not created by kpu, an error is returned)
// - create a serviceaccount token
// It then returns a copy of the currently used kubeconfig with the currently active authentication method replaced by the token.
// If minify is true, it removes all unused contexts, clusters, and authinfos from the kubeconfig.
func (k *Kubeconfig) CreateStaticTokenKubeconfig(ctx context.Context, namespace, serviceaccount, clusterrolebinding, clusterrole string, desiredValiditySeconds int64, minify bool) (*k8sapiv1.Config, time.Time, error) {
	// fetch namespace
	ns := &corev1.Namespace{}
	ns.SetName(namespace)
	if err := k.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, time.Now(), fmt.Errorf("error fetching namespace: %w", err)
		}
		// create namespace
		ns.SetLabels(map[string]string{KPUCreatedByLabel: KPUIdentity})
		if err := k.Create(ctx, ns); err != nil {
			return nil, time.Now(), fmt.Errorf("error creating namespace: %w", err)
		}
	}

	// fetch serviceaccount
	sa := &corev1.ServiceAccount{}
	sa.SetName(serviceaccount)
	sa.SetNamespace(namespace)
	if err := k.Get(ctx, client.ObjectKeyFromObject(sa), sa); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, time.Now(), fmt.Errorf("error fetching serviceaccount: %w", err)
		}
		// create serviceaccount
		sa.SetLabels(map[string]string{KPUCreatedByLabel: KPUIdentity})
		if err := k.Create(ctx, sa); err != nil {
			return nil, time.Now(), fmt.Errorf("error creating serviceaccount: %w", err)
		}
	}

	// fetch clusterrolebinding
	crb := &rbacv1.ClusterRoleBinding{}
	crb.SetName(clusterrolebinding)
	crb.SetLabels(map[string]string{KPUCreatedByLabel: KPUIdentity})
	if err := k.Get(ctx, client.ObjectKeyFromObject(crb), crb); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, time.Now(), fmt.Errorf("error fetching clusterrolebinding: %w", err)
		}
	}
	if (crb.Labels != nil) && (crb.Labels[KPUCreatedByLabel] != KPUIdentity) {
		return nil, time.Now(), fmt.Errorf("clusterrolebinding with name '%s' already exists and was not created by kpu", clusterrolebinding)
	}
	if _, err := ctrl.CreateOrUpdate(ctx, k.Client, crb, func() error {
		crb.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterrole,
		}
		crb.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceaccount,
				Namespace: namespace,
			},
		}
		return nil
	}); err != nil {
		return nil, time.Now(), fmt.Errorf("error creating/updating clusterrolebinding: %w", err)
	}

	tr, err := k.CreateServiceAccountToken(ctx, sa, desiredValiditySeconds)
	if err != nil {
		return nil, time.Now(), fmt.Errorf("error creating serviceaccount token: %w", err)
	}

	// build kubeconfig
	resInternal := k.Raw.DeepCopy()
	if minify {
		resInternal.Contexts = map[string]*k8sapi.Context{resInternal.CurrentContext: resInternal.Contexts[resInternal.CurrentContext]}
		resInternal.Clusters = map[string]*k8sapi.Cluster{resInternal.Contexts[resInternal.CurrentContext].Cluster: resInternal.Clusters[resInternal.Contexts[resInternal.CurrentContext].Cluster]}
		resInternal.AuthInfos = map[string]*k8sapi.AuthInfo{resInternal.Contexts[resInternal.CurrentContext].AuthInfo: resInternal.AuthInfos[resInternal.Contexts[resInternal.CurrentContext].AuthInfo]}
	}
	resInternal.AuthInfos[resInternal.Contexts[resInternal.CurrentContext].AuthInfo] = &k8sapi.AuthInfo{
		Token: tr.Status.Token,
	}

	res := &k8sapiv1.Config{}
	res.SetGroupVersionKind(k8sapiv1.SchemeGroupVersion.WithKind("Config"))
	if err := k8sapiv1.Convert_api_Config_To_v1_Config(resInternal, res, nil); err != nil {
		return nil, time.Now(), fmt.Errorf("error converting kubeconfig to v1: %w", err)
	}

	return res, tr.Status.ExpirationTimestamp.Time, nil
}

// CreateServiceAccountToken is a convenience function that wraps the creation of a serviceaccount token.
func (k *Kubeconfig) CreateServiceAccountToken(ctx context.Context, sa *corev1.ServiceAccount, desiredValiditySeconds int64) (*authenticationv1.TokenRequest, error) {
	token := &authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: &desiredValiditySeconds,
		},
	}
	err := k.SubResource("token").Create(ctx, sa, token)
	return token, err
}

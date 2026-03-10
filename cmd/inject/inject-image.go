package inject

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/Diaphteiros/kpu/pkg/utils"
	openmcpproviderv1alpha1 "github.com/openmcp-project/openmcp-operator/api/provider/v1alpha1"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	rtype string
	initc bool

	validResourceTypes = []string{POD, DEPLOYMENT, STATEFULSET, DAEMONSET, PLATFORMSERVICE, SERVICEPROVIDER, CLUSTERPROVIDER}
)

const (
	DEPLOYMENT      = "deployment"
	POD             = "pod"
	STATEFULSET     = "statefulset"
	DAEMONSET       = "daemonset"
	PLATFORMSERVICE = "platformservice"
	SERVICEPROVIDER = "serviceprovider"
	CLUSTERPROVIDER = "clusterprovider"
)

// InjectImageCmd represents the 'inject image' command
var InjectImageCmd = &cobra.Command{
	Use:     "image <resource-name> <image-replacement>",
	Aliases: []string{"img", "i"},
	Args:    cobra.ExactArgs(2),
	Short:   "Replace the image of a container in a pod, deployment, statefulset, or daemonset, or in one of the OpenMCP resources platformservice, serviceprovider, or clusterprovider",
	Long: `Replace the image of a container in a pod, deployment, statefulset, or daemonset, or in one of the OpenMCP resources platformservice, serviceprovider, or clusterprovider.

By default, the command targets deployments. Use the --resource-type flag to change this.
You can specify either a complete image, consisting of the repository followed by either ':' and a tag or '@' and a digest,
or just specify the tag or digest. In the latter case, the repository of the existing image will be used.
If you prefix the tag/digest with ':'/'@', the respective separator will be used, otherwise the one from the existing image will be used.

This means that these are valid formats for <image-replacement>:
- <image-repo>:<image-tag>
- <image-repo>@<image-digest>
- :<image-tag>
- @<image-digest>
- <image-tag>
- <image-digest>

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
	`,
	Run: func(cmd *cobra.Command, args []string) {
		ValidateInjectImageCommand(args)

		objName := args[0]
		rawImage := args[1]
		imageSeparator := "@"
		imageParts := strings.Split(rawImage, imageSeparator)
		if len(imageParts) == 1 {
			imageSeparator = ":"
			imageParts = strings.Split(rawImage, imageSeparator)
		}
		imageRepo := ""
		imageTag := imageParts[0]
		if len(imageParts) == 2 {
			imageRepo = imageParts[0]
			imageTag = imageParts[1]
		} else {
			// the user did not specify any separator, so we use the one from the image to be replaced
			imageSeparator = ""
		}

		// load kubeconfig
		k, err := utils.LoadKubeconfig(k8sOptions.KubeconfigPath)
		if err != nil {
			utils.Fatal(1, "error loading kubeconfig: %s", err.Error())
		}

		// get list of containers from resource
		var obj client.Object
		var containers []corev1.Container
		isOpenmcpResource := false
		switch rtype {
		case DEPLOYMENT:
			obj = &appsv1.Deployment{}
		case STATEFULSET:
			obj = &appsv1.StatefulSet{}
		case DAEMONSET:
			obj = &appsv1.DaemonSet{}
		case POD:
			obj = &corev1.Pod{}
		case PLATFORMSERVICE:
			obj = &openmcpproviderv1alpha1.PlatformService{}
			isOpenmcpResource = true
		case SERVICEPROVIDER:
			obj = &openmcpproviderv1alpha1.ServiceProvider{}
			isOpenmcpResource = true
		case CLUSTERPROVIDER:
			obj = &openmcpproviderv1alpha1.ClusterProvider{}
			isOpenmcpResource = true
		}

		obj.SetName(objName)
		if k8sOptions.Namespace != "" {
			obj.SetNamespace(k8sOptions.Namespace)
		} else {
			obj.SetNamespace(k.DefaultNamespace)
		}
		if err := k.Get(cmd.Context(), client.ObjectKeyFromObject(obj), obj); err != nil {
			utils.Fatal(1, "error getting %s: %s", rtype, err.Error())
		}

		var imageSource *string
		var replaceSource string

		if !isOpenmcpResource {
			switch rtype {
			case DEPLOYMENT:
				if initc {
					containers = obj.(*appsv1.Deployment).Spec.Template.Spec.InitContainers
				} else {
					containers = obj.(*appsv1.Deployment).Spec.Template.Spec.Containers
				}
			case STATEFULSET:
				if initc {
					containers = obj.(*appsv1.StatefulSet).Spec.Template.Spec.InitContainers
				} else {
					containers = obj.(*appsv1.StatefulSet).Spec.Template.Spec.Containers
				}
			case DAEMONSET:
				if initc {
					containers = obj.(*appsv1.DaemonSet).Spec.Template.Spec.InitContainers
				} else {
					containers = obj.(*appsv1.DaemonSet).Spec.Template.Spec.Containers
				}
			case POD:
				if initc {
					containers = obj.(*corev1.Pod).Spec.InitContainers
				} else {
					containers = obj.(*corev1.Pod).Spec.Containers
				}
			}

			if len(containers) == 0 {
				initMod := ""
				if initc {
					initMod = "init "
				}
				utils.Fatal(1, "no %scontainers found in %s", initMod, rtype)
			}

			cIdx := -1
			if len(containers) == 1 {
				cIdx = 0
			} else {
				// build prompt
				p := strings.Builder{}
				p.WriteString("Choose a container to inject the image into:\n")
				for idx, c := range containers {
					p.WriteString("  ")
					p.WriteString(fmt.Sprint(idx))
					p.WriteString(": ")
					p.WriteString(c.Name)
					p.WriteString(" - ")
					p.WriteString(c.Image)
					p.WriteString("\n")
				}
				p.WriteString("\nEnter the container's number. An empty, non-negative, non-integer, or out-of-range input will abort the command.")
				val := utils.PromptForInput(p.String(), true, nil)
				cIdx, err = strconv.Atoi(val)
				if err != nil || cIdx < 0 || cIdx >= len(containers) {
					fmt.Println("Command aborted.")
					os.Exit(0)
				}
			}
			imageSource = &containers[cIdx].Image
			replaceSource = fmt.Sprintf("container '%s' of the %s '%s'", containers[cIdx].Name, rtype, objName)
		} else {
			// openmcp resource
			var deplSpec *openmcpproviderv1alpha1.DeploymentSpec
			switch rtype {
			case PLATFORMSERVICE:
				deplSpec = &obj.(*openmcpproviderv1alpha1.PlatformService).Spec.DeploymentSpec
			case SERVICEPROVIDER:
				deplSpec = &obj.(*openmcpproviderv1alpha1.ServiceProvider).Spec.DeploymentSpec
			case CLUSTERPROVIDER:
				deplSpec = &obj.(*openmcpproviderv1alpha1.ClusterProvider).Spec.DeploymentSpec
			}
			if deplSpec == nil {
				utils.Fatal(1, "no deployment spec found in %s", rtype)
			}
			imageSource = &deplSpec.Image
			replaceSource = fmt.Sprintf("%s '%s'", rtype, objName)
		}

		// replace image
		if imageRepo == "" {
			// user specified only the version, take the current repo+
			exImgParts := strings.Split(*imageSource, "@")
			if len(exImgParts) == 2 {
				if imageSeparator == "" {
					imageSeparator = "@"
				}
			} else {
				exImgParts = strings.Split(*imageSource, ":")
				if len(exImgParts) == 2 {
					if imageSeparator == "" {
						imageSeparator = ":"
					}
				} else {
					utils.Fatal(1, "error extracting repository from image '%s'", *imageSource)
				}
			}
			imageRepo = exImgParts[0]
		}
		newImage := fmt.Sprintf("%s%s%s", imageRepo, imageSeparator, imageTag)

		if !utils.PromptForConfirmation(fmt.Sprintf("This will replace the current image '%s' with '%s' in the %s.\n", *imageSource, newImage, replaceSource), false) {
			fmt.Println("Command aborted.")
			os.Exit(0)
		}

		fmt.Printf("Replacing image '%s' with '%s' ...\n", *imageSource, newImage)
		*imageSource = newImage

		// update resource
		if err := k.Update(cmd.Context(), obj); err != nil {
			utils.Fatal(1, "error updating %s: %s", rtype, err.Error())
		}
		fmt.Printf("Image successfully injected.\n")
	},
}

func init() {
	InjectImageCmd.Flags().StringVarP(&rtype, "resource-type", "t", "deployment", fmt.Sprintf("Which resource to inject the image into. Valid values are [%s].", strings.Join(validResourceTypes, ", ")))
	InjectImageCmd.Flags().BoolVar(&initc, "init", false, "Inject the image into an init container instead of a regular container.")
}

func ValidateInjectImageCommand(args []string) {
	if !slices.Contains(validResourceTypes, rtype) {
		utils.Fatal(1, "invalid resource type '%s'", rtype)
	}
}

/*
Copyright 2019 IBM Corporation
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	kamv1 "github.com/kappnav/operator/pkg/apis/actions/v1"
	kappnavv1 "github.com/kappnav/operator/pkg/apis/kappnav/v1"
	appv1beta1 "github.com/kubernetes-sigs/application/pkg/apis/app/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// KappnavExtension extends the reconciler to manage additional resources.
type KappnavExtension interface {
	ReconcileAdditionalResources(logger Logger, request reconcile.Request, r *ReconcilerBase, instance *kappnavv1.Kappnav) (reconcile.Result, error)
}

const (
	// APIContainerName ...
	APIContainerName string = "kappnav-api"
	// UIContainerName ...
	UIContainerName string = "kappnav-ui"
	// ControllerContainerName ...
	ControllerContainerName string = "kappnav-controller"
	// OAuthProxyContainerName ...
	OAuthProxyContainerName string = "oauth-proxy"
	// OAuthProxyContainerConfigKey ...
	OAuthProxyContainerConfigKey string = "oauthProxy"
	// ServiceAccountNameSuffix ...
	ServiceAccountNameSuffix string = "sa"
)

const (
	// OAuthRedirectAnnotationName ...
	OAuthRedirectAnnotationName string = "serviceaccounts.openshift.io/oauth-redirectreference.primary"
	// OAuthVolumeName ...
	OAuthVolumeName string = "ui-service-tls"
	// OAuthVolumeMountPath ...
	OAuthVolumeMountPath string = "/etc/tls/private"
)

// GetLabels ...
func GetLabels(instance *kappnavv1.Kappnav,
	existingLabels map[string]string, component *metav1.ObjectMeta, mapType string) map[string]string {
	labels := make(map[string]string)
	labels["app.kubernetes.io/name"] = instance.Name
	labels["app.kubernetes.io/instance"] = instance.Name
	labels["app.kubernetes.io/managed-by"] = "kappnav-operator"
	if existingLabels["kappnav.io/map-type"] == ""  {
		if strings.HasSuffix(mapType, "action") {
			labels["kappnav.io/map-type"] = "action"
		} else if strings.HasSuffix(mapType, "status") {
			labels["kappnav.io/map-type"] = "status"
		} else if strings.HasSuffix(mapType, "sections") {
			labels["kappnav.io/map-type"] = "sections"
		} else if strings.HasSuffix(mapType, "builtin"){
			labels["kappnav.io/map-type"] = "builtin"
		}
	}
		
	if component != nil && len(component.Name) > 0 {
		labels["app.kubernetes.io/component"] = component.GetName()
	}
	// Allow app.kubernetes.io/name to be overriden by the CR.
	// See: https://github.com/appsody/appsody-operator/issues/179
	for key, value := range instance.Labels {
		if key != "app.kubernetes.io/instance" &&
			key != "app.kubernetes.io/component" &&
			key != "app.kubernetes.io/managed-by" {
			labels[key] = value
		}
	}
	if existingLabels == nil {
		return labels
	}
	// Add labels to the existing map.
	for key, value := range labels {
		existingLabels[key] = value
	}
	return existingLabels
}

// CustomizeServiceAccount ...
func CustomizeServiceAccount(logger Logger, sa *corev1.ServiceAccount, uiService *metav1.ObjectMeta, instance *kappnavv1.Kappnav) {
	sa.Labels = GetLabels(instance, sa.Labels, &sa.ObjectMeta, "")
	if sa.Annotations == nil {
		sa.Annotations = make(map[string]string)
	}
	// Adding the OAuth proxy to the service account.
	sa.Annotations[OAuthRedirectAnnotationName] = "{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"" + uiService.GetName() + "\"}}"
	// Adding pull secrets to the service account.
	imagePullSecrets := make([]corev1.LocalObjectReference, 1)
	imagePullSecrets[0] = corev1.LocalObjectReference{
		Name: "sa-" + sa.GetNamespace(),
	}
	pullSecrets := instance.Spec.Image.PullSecrets
	if pullSecrets != nil && len(pullSecrets) != 0 {
		for _, secretName := range pullSecrets {
			imagePullSecrets = append(imagePullSecrets, corev1.LocalObjectReference{
				Name: secretName,
			})
		}
	}
	sa.ImagePullSecrets = imagePullSecrets
}

// CustomizeClusterRoleBinding ...
func CustomizeClusterRoleBinding(crb *rbacv1.ClusterRoleBinding,
	sa *corev1.ServiceAccount, instance *kappnavv1.Kappnav) {
	crb.Labels = GetLabels(instance, crb.Labels, &crb.ObjectMeta, "")
	crb.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      sa.GetName(),
			Namespace: sa.GetNamespace(),
		},
	}
	crb.RoleRef = rbacv1.RoleRef{
		Kind:     "ClusterRole",
		Name:     "cluster-admin",
		APIGroup: "rbac.authorization.k8s.io",
	}
}

// CustomizeConfigMap ...
func CustomizeConfigMap(configMap *corev1.ConfigMap, instance *kappnavv1.Kappnav, mapType string) {
	configMap.Labels = GetLabels(instance, configMap.Labels, &configMap.ObjectMeta, mapType)
}

// CustomizeSecret ...
func CustomizeSecret(secret *corev1.Secret, instance *kappnavv1.Kappnav) {
	secret.Labels = GetLabels(instance, secret.Labels, &secret.ObjectMeta, "")
}

// CustomizeApplication ...
func CustomizeApplication(app *appv1beta1.Application, instance *kappnavv1.Kappnav, annotations map[string]string) {
	app.Labels = GetLabels(instance, app.Labels, &app.ObjectMeta, "")
	if app.Annotations == nil {
		app.Annotations = annotations
	} else {
		// Add annotations to the existing map.
		for key, value := range annotations {
			app.Annotations[key] = value
		}
	}
}

// CustomizeService ...
func CustomizeService(service *corev1.Service, instance *kappnavv1.Kappnav, annotations map[string]string) {
	service.Labels = GetLabels(instance, service.Labels, &service.ObjectMeta, "")
	if service.Annotations == nil {
		service.Annotations = annotations
	} else {
		// Add annotations to the existing map.
		for key, value := range annotations {
			service.Annotations[key] = value
		}
	}
}

// CustomizeUIServiceSpec ...
func CustomizeUIServiceSpec(serviceSpec *corev1.ServiceSpec, instance *kappnavv1.Kappnav) {
	isMinikube := IsMinikubeEnv(instance.Spec.Env.KubeEnv)
	oldType := serviceSpec.Type
	if isMinikube {
		serviceSpec.Type = corev1.ServiceTypeNodePort
	} else {
		serviceSpec.Type = ""
	}
	if oldType != serviceSpec.Type {
		serviceSpec.Ports = nil
	}
	if serviceSpec.Ports == nil || len(serviceSpec.Ports) == 0 {
		if isMinikube {
			serviceSpec.Ports = []corev1.ServicePort{
				{
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
					Protocol:   corev1.ProtocolTCP,
					Name:       "https",
				},
			}
		} else {
			serviceSpec.Ports = []corev1.ServicePort{
				{
					Name:       "proxy",
					Port:       443,
					TargetPort: intstr.FromInt(8443),
				},
			}
		}
	}
	serviceSpec.Selector = map[string]string{
		"app.kubernetes.io/component": instance.GetName() + "-ui",
	}
}

// CustomizeIngress ...
func CustomizeIngress(ingress *extensionsv1beta1.Ingress, instance *kappnavv1.Kappnav) {
	ingress.Labels = GetLabels(instance, ingress.Labels, &ingress.ObjectMeta, "")
}

// CustomizeUIIngressSpec ...
func CustomizeUIIngressSpec(ingressSpec *extensionsv1beta1.IngressSpec,
	uiService *corev1.Service, instance *kappnavv1.Kappnav) {
	if ingressSpec.Rules == nil || len(ingressSpec.Rules) == 0 {
		ingressSpec.Rules = []extensionsv1beta1.IngressRule{
			{
				IngressRuleValue: extensionsv1beta1.IngressRuleValue{
					HTTP: &extensionsv1beta1.HTTPIngressRuleValue{
						Paths: []extensionsv1beta1.HTTPIngressPath{
							{
								Path: "/kappnav-ui",
								Backend: extensionsv1beta1.IngressBackend{
									ServiceName: uiService.GetName(),
									ServicePort: intstr.FromInt(3000),
								},
							},
							{
								Path: "/kappnav",
								Backend: extensionsv1beta1.IngressBackend{
									ServiceName: uiService.GetName(),
									ServicePort: intstr.FromInt(3000),
								},
							},
						},
					},
				},
			},
		}
	}
}

// CustomizeRoute ...
func CustomizeRoute(route *routev1.Route, instance *kappnavv1.Kappnav) {
	route.Labels = GetLabels(instance, route.Labels, &route.ObjectMeta, "")
}

// CustomizeUIRouteSpec ...
func CustomizeUIRouteSpec(routeSpec *routev1.RouteSpec,
	routeName *metav1.ObjectMeta, instance *kappnavv1.Kappnav) {
	if routeSpec.TLS == nil {
		routeSpec.TLS = &routev1.TLSConfig{}
	}
	routeSpec.TLS.Termination = routev1.TLSTerminationReencrypt
	routeSpec.To.Kind = "Service"
	routeSpec.To.Name = routeName.GetName()
}

// CustomizeDeployment ...
func CustomizeDeployment(deploy *appsv1.Deployment, instance *kappnavv1.Kappnav) {
	deploy.Labels = GetLabels(instance, deploy.Labels, &deploy.ObjectMeta, "")
	// Ensure that there's at least one replica
	if deploy.Spec.Replicas == nil || *deploy.Spec.Replicas < 1 {
		one := int32(1)
		deploy.Spec.Replicas = &one
	}
	deploy.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app.kubernetes.io/component": deploy.GetName(),
		},
	}
}

// CustomizePodSpec ...
func CustomizePodSpec(pts *corev1.PodTemplateSpec, parentComponent *metav1.ObjectMeta,
	containers []corev1.Container, volumes []corev1.Volume, instance *kappnavv1.Kappnav) {
	pts.Labels = GetLabels(instance, pts.Labels, parentComponent, "")
	pts.Spec.Containers = containers
	pts.Spec.RestartPolicy = corev1.RestartPolicyAlways
	pts.Spec.ServiceAccountName = instance.GetName() + "-" + ServiceAccountNameSuffix
	pts.Spec.Volumes = volumes
	setPodSecurity(pts)
}

// CustomizeBuiltinConfigMap ...
func CustomizeBuiltinConfigMap(logger Logger, builtinConfig *corev1.ConfigMap, r *ReconcilerBase, instance *kappnavv1.Kappnav) {
	// Initialize the config map or restore values if they have been deleted.
	if builtinConfig.Data == nil {
		builtinConfig.Data = make(map[string]string)
	}
	kubeEnv := instance.Spec.Env.KubeEnv
	if IsMinikubeEnv(kubeEnv) {
		value, _ := builtinConfig.Data["openshift-console-url"]
		if len(value) == 0 {
			builtinConfig.Data["openshift-console-url"] =
				"http://127.0.0.1:8001/api/v1/namespaces/kube-system/services/http:kubernetes-dashboard:/proxy/#!"
		}
	} else if IsOpenShift(kubeEnv) {
		value, _ := builtinConfig.Data["openshift-console-url"]
		adminValue, _ := builtinConfig.Data["openshift-admin-console-url"]
		if len(value) == 0 || len(adminValue) == 0 {
			var publicURL string
			var adminPublicURL string
			if IsOCP(kubeEnv) {
				clusterInfo := getOCPClusterInfo(logger, r)
				if clusterInfo != nil {
					publicURL = clusterInfo.ConsoleBaseAddress
					adminPublicURL = publicURL
				}

			} else {
				clusterInfo := getOKDClusterInfo(logger, r)
				if clusterInfo != nil {
					publicURL = clusterInfo.ConsolePublicURL
					adminPublicURL = clusterInfo.AdminConsolePublicURL
				}
			}
			if len(value) == 0 && len(publicURL) > 0 {
				builtinConfig.Data["openshift-console-url"] = publicURL
			}
			if len(adminValue) == 0 && len(adminPublicURL) > 0 {
				builtinConfig.Data["openshift-admin-console-url"] = adminPublicURL
			}
		}
	}
	value, _ := builtinConfig.Data["liberty-problems-dashboard"]
	if len(value) == 0 {
		builtinConfig.Data["liberty-problems-dashboard"] = "Liberty-Problems-K5-20190909"
	}
	value, _ = builtinConfig.Data["liberty-traffic-dashboard"]
	if len(value) == 0 {
		builtinConfig.Data["liberty-traffic-dashboard"] = "Liberty-Traffic-K5-20190909"
	}
	value, _ = builtinConfig.Data["grafana-dashboard"]
	if len(value) == 0 {
		builtinConfig.Data["grafana-dashboard"] = "Liberty-Metrics-G5-20190521"
	}
	value, _ = builtinConfig.Data["grafana-m2-dashboard"]
	if len(value) == 0 {
		builtinConfig.Data["grafana-m2-dashboard"] = "Liberty-Metrics-M2-G5-20190521"
	}
}

// CustomizeKappnavConfigMap ...
func CustomizeKappnavConfigMap(kappnavConfig *corev1.ConfigMap, kappnavURL string, instance *kappnavv1.Kappnav) {
	// Initialize the config map or restore values if they have been deleted.
	if kappnavConfig.Data == nil {
		kappnavConfig.Data = make(map[string]string)
	}

	value, _ := kappnavConfig.Data["status-color-mapping"]
	if len(value) == 0 {
		kappnavConfig.Data["status-color-mapping"] =
			"{ \"values\": { \"Normal\": \"GREEN\", \"Completed\": \"GREEN\", \"Pending\": \"YELLOW\", \"Warning\": \"YELLOW\", \"Problem\": \"RED\", \"Failed\": \"RED\", \"Unknown\": \"GREY\", \"In Progress\": \"BLUE\"}," +
				"\"colors\": { \"GREEN\": \"#5aa700\", \"BLUE\": \"#4589ff\", \"YELLOW\": \"#B4B017\", \"RED\": \"#A74343\", \"GREY\": \"#808080\"} }"
	}
	value, _ = kappnavConfig.Data["app-status-precedence"]
	if len(value) == 0 {
		kappnavConfig.Data["app-status-precedence"] = "[ \"Failed\", \"Problem\", \"Warning\", \"Pending\", \"In Progress\", \"Unknown\", \"Normal\", \"Completed\" ]"
	}
	value, _ = kappnavConfig.Data["status-unknown"]
	if len(value) == 0 {
		kappnavConfig.Data["status-unknown"] = "Unknown"
	}
	value, _ = kappnavConfig.Data["kappnav-sa-name"]
	if len(value) == 0 {
		kappnavConfig.Data["kappnav-sa-name"] = instance.GetName() + "-" + ServiceAccountNameSuffix
	}
	if IsOpenShift(instance.Spec.Env.KubeEnv) {
		value, _ = kappnavConfig.Data["kappnav-url"]
		if len(value) == 0 && len(kappnavURL) > 0 {
			kappnavConfig.Data["kappnav-url"] = kappnavURL
		}
	}
}

// CustomizeKAM ...
func CustomizeKAM(kam *kamv1.KindActionMapping, default_kam *kamv1.KindActionMapping, instance *kappnavv1.Kappnav) {
	kam.Labels = GetLabels(instance, kam.Labels, &kam.ObjectMeta, "")
	kam.Spec = default_kam.Spec
}

// CreateUIDeploymentContainers ...
func CreateUIDeploymentContainers(existingContainers []corev1.Container, instance *kappnavv1.Kappnav) []corev1.Container {
	// Extract environment variables from existing containers.
	var apiEnv []corev1.EnvVar = nil
	var uiEnv []corev1.EnvVar = nil
	var oauthProxyEnv []corev1.EnvVar = nil
	if existingContainers != nil {
		for _, c := range existingContainers {
			switch containerName := c.Name; containerName {
			case APIContainerName:
				apiEnv = c.Env
			case UIContainerName:
				uiEnv = c.Env
			case OAuthProxyContainerName:
				oauthProxyEnv = c.Env
			}
		}
	}
	containers := []corev1.Container{
		*createContainer(APIContainerName, instance, instance.Spec.AppNavAPI, apiEnv,
			createAPIReadinessProbe(), createAPILivenessProbe(), nil, nil, nil),
		*createContainer(UIContainerName, instance, instance.Spec.AppNavUI, uiEnv,
			createUIReadinessProbe(instance), createUILiveinessProbe(instance), createUIPorts(instance), nil, nil),
	}

	if !IsMinikubeEnv(instance.Spec.Env.KubeEnv) {
		containers = append(containers, *createContainer(OAuthProxyContainerName, instance,
			instance.Spec.ExtensionContainers[OAuthProxyContainerConfigKey], oauthProxyEnv, nil, nil,
			createOAuthProxyPorts(instance), createOAuthProxyArgs(instance), createOAuthProxyVolumeMount(instance)))
	}
	return containers
}

// CreateControllerDeploymentContainers ...
func CreateControllerDeploymentContainers(existingContainers []corev1.Container, instance *kappnavv1.Kappnav) []corev1.Container {
	// Extract environment variables from existing containers.
	var apiEnv []corev1.EnvVar = nil
	var controllerEnv []corev1.EnvVar = nil
	if existingContainers != nil {
		for _, c := range existingContainers {
			switch containerName := c.Name; containerName {
			case APIContainerName:
				apiEnv = c.Env
			case ControllerContainerName:
				controllerEnv = c.Env
			}
		}
	}
	return []corev1.Container{
		*createContainer(APIContainerName, instance, instance.Spec.AppNavAPI, apiEnv,
			createAPIReadinessProbe(), createAPILivenessProbe(), nil, nil, nil),
		*createContainer(ControllerContainerName, instance, instance.Spec.AppNavController, controllerEnv,
			createControllerReadinessProbe(), createControllerLivenessProbe(), nil, nil, nil),
	}
}

// CreateUIVolumes ...
func CreateUIVolumes(instance *kappnavv1.Kappnav) []corev1.Volume {
	name := instance.Name + "-" + OAuthVolumeName
	return []corev1.Volume{
		{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: name,
				},
			},
		},
	}
}

func createContainer(name string, instance *kappnavv1.Kappnav,
	containerConfig *kappnavv1.KappnavContainerConfiguration,
	existingEnv []corev1.EnvVar,
	readinessProbe *corev1.Probe,
	livenessProbe *corev1.Probe,
	ports []corev1.ContainerPort,
	args []string,
	volumeMount *corev1.VolumeMount) *corev1.Container {
	container := &corev1.Container{
		Name:            name,
		Image:           string(containerConfig.Repository) + ":" + string(containerConfig.Tag),
		ImagePullPolicy: instance.Spec.Image.PullPolicy,
		Env: []corev1.EnvVar{
			{
				Name:  "KAPPNAV_CR_NAME",
				Value: instance.Name,
			},
			{
				Name:  "KAPPNAV_CONFIG_NAMESPACE",
				Value: instance.Namespace,
			},
			{
				Name:  "KUBE_ENV",
				Value: string(instance.Spec.Env.KubeEnv),
			},
		},
		ReadinessProbe: readinessProbe,
		LivenessProbe:  livenessProbe,
		Ports:          ports,
		Args:           args,
	}
	// Copy custom environment variable settings.
	if existingEnv != nil {
		for _, envVar := range existingEnv {
			if envVar.Name != "KAPPNAV_CR_NAME" &&
				envVar.Name != "KAPPNAV_CONFIG_NAMESPACE" &&
				envVar.Name != "KUBE_ENV" {
				container.Env = append(container.Env, envVar)
			}
		}
	}
	// Add volume mount if specified.
	if volumeMount != nil {
		container.VolumeMounts = []corev1.VolumeMount{*volumeMount}
	}
	// Apply resource constraints if enabled.
	if containerConfig.Resources.Enabled {
		container.Resources = corev1.ResourceRequirements{
			Limits:   corev1.ResourceList{},
			Requests: corev1.ResourceList{},
		}
		limits := containerConfig.Resources.Limits
		cpuLimit, err := resource.ParseQuantity(limits.CPU)
		if err == nil {
			container.Resources.Limits[corev1.ResourceCPU] = cpuLimit
		}
		memoryLimit, err := resource.ParseQuantity(limits.Memory)
		if err == nil {
			container.Resources.Limits[corev1.ResourceMemory] = memoryLimit
		}
		requests := containerConfig.Resources.Requests
		cpuRequest, err := resource.ParseQuantity(requests.CPU)
		if err == nil {
			container.Resources.Requests[corev1.ResourceCPU] = cpuRequest
		}
		memoryRequest, err := resource.ParseQuantity(requests.Memory)
		if err == nil {
			container.Resources.Requests[corev1.ResourceMemory] = memoryRequest
		}
	}
	setContainerSecurity(container)
	return container
}

func createAPIReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/kappnav/health",
				Scheme: corev1.URISchemeHTTPS,
				Port:   intstr.FromInt(9443),
			},
		},
		InitialDelaySeconds: 60,
		PeriodSeconds:       15,
		FailureThreshold:    6,
	}
}

func createAPILivenessProbe() *corev1.Probe {
	probe := createAPIReadinessProbe()
	probe.InitialDelaySeconds = 120
	return probe
}

func createUIReadinessProbe(instance *kappnavv1.Kappnav) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/health",
				Scheme: corev1.URISchemeHTTP,
				Port:   intstr.FromInt(3000),
			},
		},
		InitialDelaySeconds: 20,
		PeriodSeconds:       10,
		FailureThreshold:    6,
	}
}

func createUILiveinessProbe(instance *kappnavv1.Kappnav) *corev1.Probe {
	probe := createUIReadinessProbe(instance)
	probe.InitialDelaySeconds = 40
	probe.PeriodSeconds = 30
	return probe
}

func createUIPorts(instance *kappnavv1.Kappnav) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			ContainerPort: 3000,
			Name:          "http",
			Protocol:      corev1.ProtocolTCP,
		},
	}
}

func createOAuthProxyPorts(instance *kappnavv1.Kappnav) []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			ContainerPort: 8443,
			Name:          "public",
		},
	}
}

func createOAuthProxyArgs(instance *kappnavv1.Kappnav) []string {
	return []string{
		"--https-address=:8443",
		"--provider=openshift",
		"--openshift-service-account=" + instance.GetName() + "-" + ServiceAccountNameSuffix,
		"--upstream=http://localhost:3000",
		"--tls-cert=/etc/tls/private/tls.crt",
		"--tls-key=/etc/tls/private/tls.key",
		"--cookie-secret=SECRET",
		"--cookie-name=ssn",
		"--cookie-expire=2h",
		"--skip-provider-button=true",
		"--skip-auth-regex=.*appLauncher.js|.*featuredApp.js|.*appNavIcon.css|.*KAppNavlogo.svg",
	}
}

func createOAuthProxyVolumeMount(instance *kappnavv1.Kappnav) *corev1.VolumeMount {
	volumeMount := &corev1.VolumeMount{
		MountPath: OAuthVolumeMountPath,
		Name:      instance.Name + "-" + OAuthVolumeName,
	}
	return volumeMount
}

func createControllerReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{
					"/bin/bash",
					"-c",
					"testcntlr.sh",
				},
			},
		},
		InitialDelaySeconds: 30,
		PeriodSeconds:       5,
		FailureThreshold:    6,
	}
}

func createControllerLivenessProbe() *corev1.Probe {
	probe := createControllerReadinessProbe()
	probe.InitialDelaySeconds = 120
	probe.PeriodSeconds = 30
	return probe
}

func setContainerSecurity(container *corev1.Container) {
	f := false
	container.SecurityContext = &corev1.SecurityContext{
		Privileged:               &f,
		ReadOnlyRootFilesystem:   &f,
		AllowPrivilegeEscalation: &f,
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
	}
}

func setPodSecurity(pts *corev1.PodTemplateSpec) {
	pts.Spec.HostNetwork = false
	pts.Spec.HostPID = false
	pts.Spec.HostIPC = false
	t := true
	user := int64(1001)
	pts.Spec.SecurityContext = &corev1.PodSecurityContext{
		RunAsNonRoot: &t,
		RunAsUser:    &user,
	}
}

// IsMinikubeEnv ...
func IsMinikubeEnv(kubeEnv string) bool {
	return kubeEnv == "minikube" || kubeEnv == "k8s"
}

// IsOpenShift ...
func IsOpenShift(kubeEnv string) bool {
	return kubeEnv == "minishift" || kubeEnv == "okd" || IsOCP(kubeEnv)
}

// IsOCP ...
func IsOCP(kubeEnv string) bool {
	return kubeEnv == "ocp"
}

// OCPConsoleConfig ...
type OCPConsoleConfig struct {
	ClusterInfo OCPClusterInfo `yaml:"clusterInfo,omitempty"`
}

// OCPClusterInfo ...
type OCPClusterInfo struct {
	ConsoleBaseAddress string `yaml:"consoleBaseAddress,omitempty"`
}

func getOCPClusterInfo(logger Logger, r *ReconcilerBase) *OCPClusterInfo {
	config, err := r.GetOperatorConfigMap(logger, "console-config", "openshift-console")
	if err != nil {
		if logger.IsEnabled(LogTypeError) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Could not retrieve console-config ConfigMap for OCP, Error: %s ", err), logName)
		}
		return nil
	}
	if config.Data != nil {
		value, _ := config.Data["console-config.yaml"]
		if len(value) > 0 {
			consoleConfig := &OCPConsoleConfig{}
			err = yaml.Unmarshal([]byte(value), consoleConfig)
			if err != nil {
				if logger.IsEnabled(LogTypeError) {
					logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Could not parse console-config.yaml, Error: %s ", err), logName)
				}
				return nil
			}
			address := consoleConfig.ClusterInfo.ConsoleBaseAddress
			if len(address) > 0 {
				if strings.HasSuffix(address, "/") {
					consoleConfig.ClusterInfo.ConsoleBaseAddress = address[0 : len(address)-1]
				}
			}
			return &consoleConfig.ClusterInfo
		}
	}
	if logger.IsEnabled(LogTypeInfo) {
		logger.Log(CallerName(), LogTypeInfo, "Could not retrieve cluster info from console-config for OCP.", logName)
	}
	return nil
}

// OKDConsoleConfig ...
type OKDConsoleConfig struct {
	ClusterInfo OKDClusterInfo `yaml:"clusterInfo,omitempty"`
}

// OKDClusterInfo ...
type OKDClusterInfo struct {
	ConsolePublicURL      string `yaml:"consolePublicURL,omitempty"`
	AdminConsolePublicURL string `yaml:"adminConsolePublicURL,omitempty"`
}

func getOKDClusterInfo(logger Logger, r *ReconcilerBase) *OKDClusterInfo {
	config, err := r.GetOperatorConfigMap(logger, "webconsole-config", "openshift-web-console")
	if err != nil {
		if logger.IsEnabled(LogTypeError) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Could not retrieve webconsole-config ConfigMap for OKD, Error: %s ", err), logName)
		}
		return nil
	}
	if config.Data != nil {
		value, _ := config.Data["webconsole-config.yaml"]
		if len(value) > 0 {
			consoleConfig := &OKDConsoleConfig{}
			err = yaml.Unmarshal([]byte(value), consoleConfig)
			if err != nil {
				if logger.IsEnabled(LogTypeError) {
					logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Could not parse webconsole-config.yaml, Error: %s ", err), logName)
				}
				return nil
			}
			publicURL := consoleConfig.ClusterInfo.ConsolePublicURL
			if len(publicURL) > 0 {
				if strings.HasSuffix(publicURL, "/") {
					consoleConfig.ClusterInfo.ConsolePublicURL = publicURL[0 : len(publicURL)-1]
				}
			}
			adminPublicURL := consoleConfig.ClusterInfo.AdminConsolePublicURL
			if len(adminPublicURL) > 0 {
				if strings.HasSuffix(adminPublicURL, "/") {
					consoleConfig.ClusterInfo.AdminConsolePublicURL = adminPublicURL[0 : len(adminPublicURL)-1]
				}
			}
			return &consoleConfig.ClusterInfo
		}
	}
	if logger.IsEnabled(LogTypeInfo) {
		logger.Log(CallerName(), LogTypeInfo, "Could not retrieve cluster info from webconsole-config for OKD.", logName)
	}
	return nil
}

//
// Functions for accessing and updating status on the CR
//

// GetCondition ...
func GetCondition(conditionType kappnavv1.StatusConditionType, status *kappnavv1.KappnavStatus) *kappnavv1.StatusCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return &status.Conditions[i]
		}
	}
	return nil
}

// SetCondition ...
func SetCondition(condition kappnavv1.StatusCondition, status *kappnavv1.KappnavStatus) {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condition.Type {
			status.Conditions[i] = condition
			return
		}
	}
	status.Conditions = append(status.Conditions, condition)
}

// CallerName get the caller program file name, line number and function name in "fileName:line# funcName"
func CallerName() string {
	var callerName string
	pc, fileName, line, _ := runtime.Caller(1)

	// get function name
	funcNameFull := runtime.FuncForPC(pc).Name()
	funcNameEnd := filepath.Ext(funcNameFull)
	funcName := strings.TrimPrefix(funcNameEnd, ".")

	// get file name
	suffix := ".go"
	_, nf := filepath.Split(fileName)
	if strings.HasSuffix(nf, ".go") {
		fileName = strings.TrimSuffix(nf, suffix)
		callerName = fileName + suffix + ":" + strconv.Itoa(line) + " " + funcName
	}
	return callerName
}

//ErrorWithStack print stack trace for error message
func ErrorWithStack(msg string) string {
	cause := errors.New(msg)
	err := errors.WithStack(cause)
	s := fmt.Sprintf("%+v", err)
	return s
}

//FormatTimeStamp format with unix seconds in float
func FormatTimestamp(t time.Time) float64 {
	s := fmt.Sprintf("%10.7f", float64(t.UnixNano())/1e9)
	ts, _ := strconv.ParseFloat(s, 64)
	return ts
}

/*
func contains(arr []string, corev1.LocalObjectReference) bool {
	for _, a := range arr {
	   if a == str {
		  return true
	   }
	}
	return false
 } */

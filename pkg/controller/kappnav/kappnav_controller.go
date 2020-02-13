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

package kappnav

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"fmt"

	kappnavv1 "github.com/kappnav/operator/pkg/apis/kappnav/v1"
	kappnavutils "github.com/kappnav/operator/pkg/utils"
	appv1beta1 "github.com/kubernetes-sigs/application/pkg/apis/app/v1beta1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/yaml"
)

var logName = "controller_kappnav"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Kappnav Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	logger := kappnavutils.NewLogger(true)  //log in json format 

	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Creating a new kappnav controller and adds it to the manager", logName)
	}
	return add(logger, mgr, newReconciler(logger, mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(logger kappnavutils.Logger, mgr manager.Manager) reconcile.Reconciler {
	reconciler := &ReconcileKappnav{ReconcilerBase: kappnavutils.NewReconcilerBase(mgr.GetClient(),
		mgr.GetScheme(), mgr.GetConfig(), mgr.GetRecorder("kappnav-operator"))}

	// Create CRDs if they do not already exist.
	//test
	files, err := ioutil.ReadDir("crds")
	if err != nil {
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {			
            logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to read directory: crds, error: %s", err), logName)
        }
		os.Exit(1)
	}
	for _, file := range files {
		if !file.IsDir() {
			fileName := "crds/" + file.Name()
			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				if (logger.IsEnabled(kappnavutils.LogTypeDebug)) {
					logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeDebug, fmt.Sprintf("Read file from the image: ", fileName), logName)
				}
				// Read the file from the image.
				fData, err := ioutil.ReadFile(fileName)
				if err != nil {					
					if (logger.IsEnabled(kappnavutils.LogTypeError)) {
						logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to read file: %s, error: %s", fileName, err), logName)
					}
					os.Exit(1)
				}
				crd := &apiextensionsv1beta1.CustomResourceDefinition{}
				// Unmarshal the YAML into an object.
				err = yaml.Unmarshal(fData, crd)
				if err != nil {
					if (logger.IsEnabled(kappnavutils.LogTypeError)) {
						logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to unmarshal YAML file: %s, error: %s", fileName, err), logName)
					}
					os.Exit(1)
				}
				// Create the CRD if it does not already exist.
				err = reconciler.GetClient().Create(context.TODO(), crd)
				if err != nil && !errors.IsAlreadyExists(err) {
					if (logger.IsEnabled(kappnavutils.LogTypeError)) {
						logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to create CRD: %s, error: %s", crd.GetName(), err), logName)
					}
					os.Exit(1)
				}
			}
		}
	}

	return reconciler
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(logger kappnavutils.Logger, mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kappnav-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Kappnav
	err = c.Watch(&source.Kind{Type: &kappnavv1.Kappnav{}}, &handler.EnqueueRequestForObject{})
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Watch for changes to primary resource", logName)
	}
		
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources that are always created by the operator
	// (such as Deployment, ConfigMap, Service, etc...) and requeue the owner Kappnav
	types := []runtime.Object{&appsv1.Deployment{}, &corev1.ConfigMap{}, &corev1.Secret{},
		&corev1.Service{}, &corev1.ServiceAccount{}, &rbacv1.ClusterRoleBinding{}, &appv1beta1.Application{}}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Watch for changes to secondary resource", logName)
	}	
	for i := range types {	
		err = c.Watch(&source.Kind{Type: types[i]}, &handler.EnqueueRequestForOwner{			
			IsController: true,
			OwnerType:    &kappnavv1.Kappnav{},
		})
		if err != nil {
			return err
		}
	}

	// Watch for changes to secondary resources Ingress and Route
	// (when available) and requeue the owner Kappnav
	types = []runtime.Object{&extensionsv1beta1.Ingress{}, &routev1.Route{}}
	for i := range types {
		_ = c.Watch(&source.Kind{Type: types[i]}, &handler.EnqueueRequestForOwner{			
			IsController: true,
			OwnerType:    &kappnavv1.Kappnav{},
		})
	}
	return nil
}

// blank assignment to verify that ReconcileKappnav implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKappnav{}

// ReconcileKappnav reconciles a Kappnav object
type ReconcileKappnav struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	kappnavutils.ReconcilerBase
}

// Reconcile reads that state of the cluster for a Kappnav object and makes changes based on the state read
// and what is in the Kappnav.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKappnav) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := kappnavutils.NewLogger(true) //log in json format
	var otherLogData = " in Request.Namespace: " + request.Namespace + ", Request.Name: " + request.Name

	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Reconciling Kappnav" + otherLogData, logName)		
	}

	// Fetch the Kappnav instance
	instance := &kappnavv1.Kappnav{}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {		
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Fetch kappnav instance" + otherLogData, logName)
	} 	

	err := r.GetClient().Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	
	// Call factory method to create new KappnavExtension
	extension := kappnavutils.NewKappnavExtension()

	// Apply defaults to the Kappnav instance
	err = kappnavutils.SetKappnavDefaults(instance)
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Apply defaults to kappnav instance" + otherLogData, logName)
	}
	if err != nil {
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to process default values file" + otherLogData + ", Error: %s", err), logName)
		}
		return reconcile.Result{}, err
	}

	// Retrieve logging info from kappnav CR and update the kappnavutils level	
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Retrieve kappnavutils info from kappnav CR" + otherLogData, logName)
	} 	
	loggingMap := instance.Spec.Logging	
	if len(loggingMap["operator"]) > 0 {	
		if (logger.IsEnabled(kappnavutils.LogTypeDebug)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeDebug, fmt.Sprintf("Set the log level to ", loggingMap["operator"]) + otherLogData, logName)
		} 	
		setLoggingLevel(logger, loggingMap["operator"])		
	}
	
	kappnavName := &metav1.ObjectMeta{
        Name:      "kappnav",
        Namespace: instance.GetNamespace(),
	}

	uiServiceAndRouteName := &metav1.ObjectMeta{
		Name:      instance.GetName() + "-ui-service",
		Namespace: instance.GetNamespace(),
	}
	
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-" + kappnavutils.ServiceAccountNameSuffix,
			Namespace: instance.GetNamespace(),
		},
	}
	// Create or update service account
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update service account" + otherLogData, logName)
	} 	
	err = r.CreateOrUpdate(logger, serviceAccount, instance, func() error {
		kappnavutils.CustomizeServiceAccount(logger, serviceAccount, uiServiceAndRouteName, instance)
		return nil
	})
	if err != nil {
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the ServiceAccount" + otherLogData + ", Error: %s ", err), logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}

	// Create or update cluster role binding
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-" + instance.GetNamespace() + "-crb",
			Namespace: instance.GetNamespace(),
		},
	}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update cluster role binding" + otherLogData, logName)
	} 	
	err = r.CreateOrUpdate(logger, crb, instance, func() error {
		kappnavutils.CustomizeClusterRoleBinding(crb, serviceAccount, instance)
		return nil
	})
	if err != nil && !errors.IsAlreadyExists(err) {	
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the ClusterRoleBinding" + otherLogData + ", Error: %s ", err) , logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}

	// Dummy secret for Minikube support
	dummySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-" + kappnavutils.OAuthVolumeName,
			Namespace: instance.GetNamespace(),
		},
	}

	// The kappnav application
	kappnavCR := &appv1beta1.Application{
		ObjectMeta: *kappnavName,
	}
	kappnavCRAnnotations := map[string]string{
		"kappnav.application.hidden": "true",
	}

	// Create or update the kappnav Application
	err = r.CreateOrUpdate(logger, kappnavCR, instance, func() error {
		kappnavutils.CustomizeApplication(kappnavCR, instance, kappnavCRAnnotations)
		return nil
	})
	if err != nil {
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the kappnav Application" + otherLogData + ", Error: %s ", err), logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}

	// The UI service
	uiService := &corev1.Service{
		ObjectMeta: *uiServiceAndRouteName,
	}
	uiServiceAnnotations := map[string]string{
		"service.alpha.openshift.io/serving-cert-secret-name": dummySecret.Name,
	}

	// Kappnav URL is computed from the route
	kappnavURL := ""

	isMinikube := kappnavutils.IsMinikubeEnv(instance.Spec.Env.KubeEnv)
	if isMinikube {
		// Create or update dummy secret
		err = r.CreateOrUpdate(logger, dummySecret, instance, func() error {
			kappnavutils.CustomizeSecret(dummySecret, instance)
			return nil
		})
		if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update dummy secret" + otherLogData, logName)
		} 	
		if err != nil {			
			if (logger.IsEnabled(kappnavutils.LogTypeError)) {
				logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the dummy secret" + otherLogData + ", Error: %s ", err), logName)
			}
			return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
		}
		// Create or update the UI service
		err = r.CreateOrUpdate(logger, uiService, instance, func() error {
			kappnavutils.CustomizeService(uiService, instance, uiServiceAnnotations)
			kappnavutils.CustomizeUIServiceSpec(&uiService.Spec, instance)
			return nil
		})
		if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update UI service" + otherLogData, logName)
		} 	
		if err != nil {			
			if (logger.IsEnabled(kappnavutils.LogTypeError)) {
				logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the UI service" + otherLogData + ", Error: %s ", err), logName)
			}
			return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
		}
		// Create or update UI ingress
		uiIngress := &extensionsv1beta1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.GetName() + "-ui-ingress",
				Namespace: instance.GetNamespace(),
			},
		}
		if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update UI ingress" + otherLogData, logName)
		} 	
		err = r.CreateOrUpdate(logger, uiIngress, instance, func() error {
			kappnavutils.CustomizeIngress(uiIngress, instance)
			kappnavutils.CustomizeUIIngressSpec(&uiIngress.Spec, uiService, instance)
			return nil
		})
		if err != nil {			
			if (logger.IsEnabled(kappnavutils.LogTypeError)) {
				logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the UI ingress" + otherLogData + ", Error: %s ", err), logName)
			}
			return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
		}
	} else {
		// Create or update the UI service
		err = r.CreateOrUpdate(logger, uiService, instance, func() error {
			kappnavutils.CustomizeService(uiService, instance, uiServiceAnnotations)
			kappnavutils.CustomizeUIServiceSpec(&uiService.Spec, instance)
			return nil
		})
		if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update UI service" + otherLogData, logName)
		} 	
		if err != nil {			
			if (logger.IsEnabled(kappnavutils.LogTypeError)) {
				logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the UI Service" + otherLogData + ", Error: %s ", err), logName)
			}
			return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
		}
		// Create or update UI route
		uiRoute := &routev1.Route{
			ObjectMeta: *uiServiceAndRouteName,
		}
		err = r.CreateOrUpdate(logger, uiRoute, instance, func() error {
			kappnavutils.CustomizeRoute(uiRoute, instance)
			kappnavutils.CustomizeUIRouteSpec(&uiRoute.Spec, uiServiceAndRouteName, instance)
			return nil
		})
		if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update UI route" + otherLogData, logName)
		} 	
		if err != nil {			
			if (logger.IsEnabled(kappnavutils.LogTypeError)) {
				logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the UI route" + otherLogData + ", Error: %s ", err), logName)
			}
			return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
		}
		// Compute Kappnav URL from route.
		routeHost := uiRoute.Spec.Host
		routePath := uiRoute.Spec.Path
		if len(routeHost) > 0 && len(routePath) > 0 {
			kappnavURL = "https://" + routeHost + routePath
		}
	}

	// Create or update action, section and status config maps.
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update action, section and status config maps" + otherLogData, logName)
	} 	
	mapDirs := []string{"maps/action", "maps/sections", "maps/status"}
	for _, dir := range mapDirs {
		if (logger.IsEnabled(kappnavutils.LogTypeDebug)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeDebug, "Read dir: " + dir + otherLogData, logName)
		} 	
		files, err := ioutil.ReadDir(dir)
		if err != nil {			
			if (logger.IsEnabled(kappnavutils.LogTypeError)) {
				logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the read directory %s " + otherLogData + ", Error: %s ", dir, err), logName)
			}
			return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
		}
		for _, file := range files {
			if !file.IsDir() {
				fileName := dir + "/" + file.Name()
				if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
					if (logger.IsEnabled(kappnavutils.LogTypeDebug)) {
						logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeDebug, "Read file: " + fileName + otherLogData, logName)
					} 	
					// Read the file from the image.
					fData, err := ioutil.ReadFile(fileName)
					if err != nil {						
						if (logger.IsEnabled(kappnavutils.LogTypeError)) {
							logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to read file: %s " + otherLogData + ", Error: %s ", fileName, err), logName)
						}
						return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
					}
					// Parse the file into a template.
					t, err := template.New(fileName).Parse(string(fData))
					if err != nil {						
						if (logger.IsEnabled(kappnavutils.LogTypeError)) {
							logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to parse template file: %s " + otherLogData + ", Error: %s ", fileName, err), logName)
						}
						return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
					}
					// Execute the template against the Kappnav CR instance.
					var buf bytes.Buffer
					err = t.Execute(&buf, instance)
					if err != nil {						
						if (logger.IsEnabled(kappnavutils.LogTypeError)) {
							logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to execute template: %s " + otherLogData + ", Error: %s ", fileName, err), logName)
						}
						return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
					}
					configMap := &corev1.ConfigMap{}
					// Unmarshal the YAML into an object.
					err = yaml.Unmarshal(buf.Bytes(), configMap)
					if err != nil {
						if (logger.IsEnabled(kappnavutils.LogTypeError)) {
							logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to unmarshal YAML file: %s " + otherLogData + ", Error: %s", fileName, err), logName)
						}
						return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
					}
					clusterMap := &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      configMap.GetName(),
							Namespace: instance.GetNamespace(),
						},
					}
					// Write the data to the map in the cluster.
					err = r.CreateOrUpdate(logger, clusterMap, instance, func() error {
						kappnavutils.CustomizeConfigMap(clusterMap, instance)
						// Write the data section if it doesn't exist or is empty.
						if clusterMap.Data == nil || len(clusterMap.Data) == 0 {
							clusterMap.Data = configMap.Data
						}
						return nil
					})
					if err != nil {						
						if (logger.IsEnabled(kappnavutils.LogTypeError)) {
							logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the %s ConfigMap" + otherLogData + ", Error: %s", configMap.GetName(), err), logName)
						}
						return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
					}
				}
			}
		}
	}

	// Create or update builtin config
	builtinConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "builtin",
			Namespace: instance.GetNamespace(),
		},
	}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update builtin config" + otherLogData, logName)
	} 	
	err = r.CreateOrUpdate(logger, builtinConfig, instance, func() error {
		kappnavutils.CustomizeConfigMap(builtinConfig, instance)
		kappnavutils.CustomizeBuiltinConfigMap(logger, builtinConfig, &r.ReconcilerBase, instance)
		return nil
	})
	if err != nil {		
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the kappnav-config ConfigMap" + otherLogData + ", Error: %s ", err), logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}

	// Create or update kappnav-config
	kappnavConfig := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kappnav-config",
			Namespace: instance.GetNamespace(),
		},
	}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update kappnav-config" + otherLogData, logName)
	} 	
	err = r.CreateOrUpdate(logger, kappnavConfig, instance, func() error {
		kappnavutils.CustomizeConfigMap(kappnavConfig, instance)
		kappnavutils.CustomizeKappnavConfigMap(kappnavConfig, kappnavURL, instance)
		return nil
	})
	if err != nil {		
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the kappnav-config ConfigMap" + otherLogData + ", Error: %s", err), logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}

	// Create or update the UI deployment
	uiDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-ui",
			Namespace: instance.GetNamespace(),
		},
	}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update UI deployment" + otherLogData, logName)
	} 	
	err = r.CreateOrUpdate(logger, uiDeployment, instance, func() error {
		pts := &uiDeployment.Spec.Template
		kappnavutils.CustomizeDeployment(uiDeployment, instance)
		kappnavutils.CustomizePodSpec(pts, &uiDeployment.ObjectMeta,
			kappnavutils.CreateUIDeploymentContainers(pts.Spec.Containers, instance),
			kappnavutils.CreateUIVolumes(instance), instance)
		return nil
	})
	if err != nil {
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the UI Deployment" + otherLogData + ", Error: %s ", err), logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}

	// Create or update the Controller deployment
	controllerDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.GetName() + "-controller",
			Namespace: instance.GetNamespace(),
		},
	}
	if (logger.IsEnabled(kappnavutils.LogTypeInfo)) {
		logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeInfo, "Create or update controller deployment" + otherLogData, logName)
	} 	
	err = r.CreateOrUpdate(logger, controllerDeployment, instance, func() error {
		pts := &controllerDeployment.Spec.Template
		kappnavutils.CustomizeDeployment(controllerDeployment, instance)
		kappnavutils.CustomizePodSpec(pts, &controllerDeployment.ObjectMeta,
			kappnavutils.CreateControllerDeploymentContainers(pts.Spec.Containers, instance), nil, instance)
		return nil
	})
	if err != nil {
		if (logger.IsEnabled(kappnavutils.LogTypeError)) {
			logger.Log(kappnavutils.CallerName(), kappnavutils.LogTypeError, fmt.Sprintf("Failed to reconcile the Controller Deployment" + otherLogData + ", Error: %s", err), logName)
		}
		return r.ManageError(logger, err, kappnavv1.StatusConditionTypeReconciled, instance)
	}	

	// If an extension exists call its reconcile function, otherwise return success.
	if extension != nil {
		return extension.ReconcileAdditionalResources(logger, request, &r.ReconcilerBase, instance)
	}
	return r.ManageSuccess(logger, kappnavv1.StatusConditionTypeReconciled, instance)

	
}


// get logging info from CR and set logger level
func setLoggingLevel(logger kappnavutils.Logger, loginfo string) {
	switch loginfo { 
	case "info":
		logger.SetLogLevel(kappnavutils.LogLevelInfo)
		break
	case "debug":
		logger.SetLogLevel(kappnavutils.LogLevelDebug)
		break
	case "error":
		logger.SetLogLevel(kappnavutils.LogLevelError)
		break
	case "warning":
		logger.SetLogLevel(kappnavutils.LogLevelWarning)
		break
	case "entry":
		logger.SetLogLevel(kappnavutils.LogLevelEntry)
		break
	case "all":
		logger.SetLogLevel(kappnavutils.LogLevelAll)
		break
	case "none":
		logger.SetLogLevel(kappnavutils.LogLevelNone)
		break
	}	
}




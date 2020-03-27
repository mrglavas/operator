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
	"context"
	"fmt"
	"math"
	"time"

	kappnavv1 "github.com/kappnav/operator/pkg/apis/kappnav/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

var logName = "utils"

// ReconcilerBase base reconciler with some common behaviour
type ReconcilerBase struct {
	client     client.Client
	scheme     *runtime.Scheme
	recorder   record.EventRecorder
	restConfig *rest.Config
	discovery  discovery.DiscoveryInterface
}

// NewReconcilerBase creates a new ReconcilerBase
func NewReconcilerBase(client client.Client, scheme *runtime.Scheme, restConfig *rest.Config, recorder record.EventRecorder) ReconcilerBase {
	return ReconcilerBase{
		client:     client,
		scheme:     scheme,
		recorder:   recorder,
		restConfig: restConfig,
	}
}

// GetClient returns client
func (r *ReconcilerBase) GetClient() client.Client {
	return r.client
}

// GetScheme retuns scheme
func (r *ReconcilerBase) GetScheme() *runtime.Scheme {
	return r.scheme
}

// GetRecorder returns the underlying recorder
func (r *ReconcilerBase) GetRecorder() record.EventRecorder {
	return r.recorder
}

// GetDiscoveryClient ...
func (r *ReconcilerBase) GetDiscoveryClient() (discovery.DiscoveryInterface, error) {
	if r.discovery == nil {
		var err error
		r.discovery, err = discovery.NewDiscoveryClientForConfig(r.restConfig)
		return r.discovery, err
	}

	return r.discovery, nil
}

// SetDiscoveryClient ...
func (r *ReconcilerBase) SetDiscoveryClient(discovery discovery.DiscoveryInterface) {
	r.discovery = discovery
}

// CreateOrUpdate ...
func (r *ReconcilerBase) CreateOrUpdate(logger Logger, obj metav1.Object, owner metav1.Object, reconcile func() error) error {
	mutate := func(o runtime.Object) error {
		err := reconcile()
		return err
	}

	controllerutil.SetControllerReference(owner, obj, r.scheme)
	runtimeObj, ok := obj.(runtime.Object)
	if !ok {
		err := fmt.Errorf("%T is not a runtime.Object", obj)
		if (logger.IsEnabled(LogTypeError)) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Failed to convert into runtime.Object, Error: %s ", err), logName)
		}
		return err
	}
	result, err := controllerutil.CreateOrUpdate(context.TODO(), r.GetClient(), runtimeObj, mutate)
	if err != nil {
		return err
	}

	var gvk schema.GroupVersionKind
	gvk, err = apiutil.GVKForObject(runtimeObj, r.scheme)
	if err == nil {
		if (logger.IsEnabled(LogTypeInfo)) {
			logger.Log(CallerName(), LogTypeInfo, fmt.Sprintf("Reconciled, Kind: %s, Name: %s, Status: %s ", gvk.Kind, obj.GetName(), result), logName)
		}			
	}

	return err
}

// DeleteResource deletes kubernetes resource
func (r *ReconcilerBase) DeleteResource(obj runtime.Object) error {
	logger := NewLogger(true) //log in JSON format

	err := r.client.Delete(context.TODO(), obj)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			if (logger.IsEnabled(LogTypeError)) {
				logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Unable to delete object: %s, Error: %s ", obj, err), logName)
			}
			return err
		}
		return nil
	}

	metaObj, ok := obj.(metav1.Object)
	if !ok {
		err := fmt.Errorf("%T is not a runtime.Object", obj)
		if (logger.IsEnabled(LogTypeError)) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Failed to convert into runtime.Object, Error: %s ", err), logName)
		}
		return err
	}

	var gvk schema.GroupVersionKind
	gvk, err = apiutil.GVKForObject(obj, r.scheme)
	if err == nil {
		if (logger.IsEnabled(LogTypeInfo)) {
			logger.Log(CallerName(), LogTypeInfo, fmt.Sprintf("Reconciled, Kind: %s, Name: %s, Status: deleted", gvk.Kind, metaObj.GetName()), logName)
		}
	}
	return nil
}

// DeleteResources ...
func (r *ReconcilerBase) DeleteResources(resources []runtime.Object) error {
	for i := range resources {
		err := r.DeleteResource(resources[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// GetOperatorConfigMap ...
func (r *ReconcilerBase) GetOperatorConfigMap(logger Logger, name string, ns string) (*corev1.ConfigMap, error) {
	if (logger.IsEnabled(LogTypeInfo)) {
		logger.Log(CallerName(), LogTypeInfo, fmt.Sprintf("Attempting to read ConfigMap, name: %s, namespace: %s", name, ns), logName)
	}
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Kind:    "ConfigMap",
		Version: "v1",
	})
	err := r.GetClient().Get(context.TODO(), client.ObjectKey{
		Namespace: ns,
		Name:      name,
	}, u)
	if err != nil {
		return nil, err
	}
	b, err := u.MarshalJSON()
	if err != nil {
		return nil, err
	}
	b, err = yaml.JSONToYAML(b)
	if err != nil {
		return nil, err
	}
	configMap := &corev1.ConfigMap{}
	err = yaml.Unmarshal(b, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

// ManageError ...
func (r *ReconcilerBase) ManageError(logger Logger, issue error, conditionType kappnavv1.StatusConditionType, cr *kappnavv1.Kappnav) (reconcile.Result, error) {
	r.GetRecorder().Event(cr, "Warning", "ProcessingError", issue.Error())

	oldCondition := GetCondition(conditionType, &cr.Status)
	if oldCondition == nil {
		oldCondition = &kappnavv1.StatusCondition{LastUpdateTime: metav1.Time{}}
	}

	lastUpdate := oldCondition.LastUpdateTime.Time
	lastStatus := oldCondition.Status

	// Keep the old `LastTransitionTime` when status has not changed
	nowTime := metav1.Now()
	transitionTime := oldCondition.LastTransitionTime
	if lastStatus == corev1.ConditionTrue {
		transitionTime = &nowTime
	}

	newCondition := kappnavv1.StatusCondition{
		LastTransitionTime: transitionTime,
		LastUpdateTime:     nowTime,
		Reason:             string(apierrors.ReasonForError(issue)),
		Type:               conditionType,
		Message:            issue.Error(),
		Status:             corev1.ConditionFalse,
	}

	SetCondition(newCondition, &cr.Status)

	err := r.GetClient().Status().Update(context.Background(), cr)
	if err != nil {
		if (logger.IsEnabled(LogTypeError)) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Unable to update status, Error: %s ", err), logName)
		}
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}

	// StatusReasonInvalid means the requested create or update operation cannot be
	// completed due to invalid data provided as part of the request. Don't retry.
	if apierrors.IsInvalid(issue) {
		return reconcile.Result{}, nil
	}

	var retryInterval time.Duration
	if lastUpdate.IsZero() || lastStatus == corev1.ConditionTrue {
		retryInterval = time.Second
	} else {
		retryInterval = newCondition.LastUpdateTime.Sub(lastUpdate).Round(time.Second)
	}

	return reconcile.Result{
		RequeueAfter: time.Duration(math.Min(float64(retryInterval.Nanoseconds()*2), float64(time.Hour.Nanoseconds()*6))),
		Requeue:      true,
	}, nil
}

// ManageSuccess ...
func (r *ReconcilerBase) ManageSuccess(logger Logger, conditionType kappnavv1.StatusConditionType, cr *kappnavv1.Kappnav) (reconcile.Result, error) {
	
	oldCondition := GetCondition(conditionType, &cr.Status)
	if oldCondition == nil {
		oldCondition = &kappnavv1.StatusCondition{LastUpdateTime: metav1.Time{}}
	}

	// Keep the old `LastTransitionTime` when status has not changed
	nowTime := metav1.Now()
	transitionTime := oldCondition.LastTransitionTime
	if oldCondition.Status == corev1.ConditionFalse {
		transitionTime = &nowTime
	}

	statusCondition := kappnavv1.StatusCondition{
		LastTransitionTime: transitionTime,
		LastUpdateTime:     nowTime,
		Type:               conditionType,
		Reason:             "",
		Message:            "",
		Status:             corev1.ConditionTrue,
	}

	SetCondition(statusCondition, &cr.Status)
	err := r.GetClient().Status().Update(context.Background(), cr)
	if err != nil {
		if (logger.IsEnabled(LogTypeError)) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Unable to update status, Error: %s ", err), logName)
		}
		return reconcile.Result{
			RequeueAfter: time.Second,
			Requeue:      true,
		}, nil
	}
	return reconcile.Result{}, nil
}

// IsGroupVersionSupported ...
func (r *ReconcilerBase) IsGroupVersionSupported(groupVersion string) (bool, error) {
	logger := NewLogger(true)  //log in JSON format

	cli, err := r.GetDiscoveryClient()
	if err != nil {
		if (logger.IsEnabled(LogTypeError)) {
			logger.Log(CallerName(), LogTypeError, fmt.Sprintf("Failed to return a discovery client for the current reconciler, Error: %s ", err), logName)
		}
		return false, err
	}

	_, err = cli.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

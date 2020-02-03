/*
Copyright 2019, 2020 IBM Corporation
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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KappnavSpec defines the desired state of Kappnav
// +k8s:openapi-gen=true
type KappnavSpec struct {
	AppNavAPI           *KappnavContainerConfiguration            `json:"appNavAPI,omitempty"`
	AppNavController    *KappnavContainerConfiguration            `json:"appNavController,omitempty"`
	AppNavUI            *KappnavContainerConfiguration            `json:"appNavUI,omitempty"`
	ExtensionContainers map[string]*KappnavContainerConfiguration `json:"extensionContainers,omitempty"`
	Image               *KappnavImageConfiguration                `json:"image,omitempty"`
	Env                 *Environment                              `json:"env,omitempty"`
	Logging             map[string]string                         `json:"logging,omitempty"`
}

// KappnavContainerConfiguration defines the configuration for a Kappnav container
type KappnavContainerConfiguration struct {
	Repository Repository                  `json:"repository,omitempty"`
	Tag        Tag                         `json:"tag,omitempty"`
	Resources  *KappnavResourceConstraints `json:"resources,omitempty"`
}

// KappnavResourceConstraints defines resource constraints for a Kappnav container
type KappnavResourceConstraints struct {
	Enabled  bool       `json:"enabled,omitempty"`
	Requests *Resources `json:"requests,omitempty"`
	Limits   *Resources `json:"limits,omitempty"`
}

// Resources ...
type Resources struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
}

// Repository ...
type Repository string

// Tag ...
type Tag string

// KappnavImageConfiguration ...
type KappnavImageConfiguration struct {
	PullPolicy  corev1.PullPolicy `json:"pullPolicy,omitempty"`
	PullSecrets []string          `json:"pullSecrets,omitempty"`
}

// Environment variables.
type Environment struct {
	KubeEnv string `json:"kubeEnv,omitempty"`
}

// KappnavStatus defines the observed state of Kappnav
// +k8s:openapi-gen=true
type KappnavStatus struct {
	Conditions []StatusCondition `json:"conditions,omitempty"`
}

// StatusCondition ...
type StatusCondition struct {
	LastTransitionTime *metav1.Time           `json:"lastTransitionTime,omitempty"`
	LastUpdateTime     metav1.Time            `json:"lastUpdateTime,omitempty"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	Status             corev1.ConditionStatus `json:"status,omitempty"`
	Type               StatusConditionType    `json:"type,omitempty"`
}

// StatusConditionType ...
type StatusConditionType string

const (
	// StatusConditionTypeReconciled ...
	StatusConditionTypeReconciled StatusConditionType = "Reconciled"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Kappnav is the Schema for the kappnavs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Kappnav struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KappnavSpec   `json:"spec,omitempty"`
	Status KappnavStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KappnavList contains a list of Kappnav
type KappnavList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kappnav `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kappnav{}, &KappnavList{})
}

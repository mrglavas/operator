package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KindActionMappingSpec defines the desired state of KindActionMapping
// +k8s:openapi-gen=true
type KindActionMappingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Precedence int                    `json:"precedence,omitempty"`
	Mappings   []MappingConfiguration `json:"mappings,omitempty"`
}

// MappingsConfigurtion defines resource constraints for Mapping configuration
type MappingConfiguration struct {
	ApiVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Subkind    string `json:"subkind,omitempty"`
	Name       string `json:"name,omitempty"`
	Mapname    string `json:"mapname,omitempty"`
}

// KindActionMappingStatus defines the observed state of KindActionMapping
// +k8s:openapi-gen=true
type KindActionMappingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KindActionMapping is the Schema for the kindactionmappings API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type KindActionMapping struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KindActionMappingSpec   `json:"spec,omitempty"`
	Status KindActionMappingStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KindActionMappingList contains a list of KindActionMapping
type KindActionMappingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KindActionMapping `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KindActionMapping{}, &KindActionMappingList{})
}

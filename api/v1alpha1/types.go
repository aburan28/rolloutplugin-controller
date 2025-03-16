// +kubebuilder:object:generate=true
// +groupName=rolloutsplugin.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	myscheme := runtime.NewScheme()
	SchemeBuilder.Register(&RolloutPlugin{}, &RolloutPluginList{})
	utilruntime.Must(AddToScheme(myscheme))
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=rolloutplugins,scope=Namespaced
// +kubebuilder:subresource:status
type RolloutPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RolloutPluginSpec   `json:"spec,omitempty"`
	Status RolloutPluginStatus `json:"status,omitempty"`
}

type RolloutPluginSpec struct {
	Plugin   Plugin   `json:"plugin"`
	Strategy Strategy `json:"strategy"`
}

type Strategy struct {
	Type   string `json:"type"`
	Canary Canary `json:"canary"`
}

type Canary struct {
	Steps []Step `json:"steps"`
}

type Step struct {
	Step string `json:"step"`
}

type Plugin struct {
	Name   string `json:"name"`
	Sha256 string `json:"sha256"`
	Url    string `json:"url"`
}

type RolloutPluginStatus struct {
	Conditions         []Condition `json:"conditions"`
	Initialized        bool        `json:"initialized"`
	ObservedGeneration int64       `json:"observedGeneration"`
}

type Condition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastUpdateTime     metav1.Time `json:"lastUpdateTime" protobuf:"bytes,3,opt,name=lastUpdateTime"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	Reason             Reason      `json:"reason"`
	Message            string      `json:"message"`
}

type Reason string

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=rolloutpluginlist,scope=Namespaced
// +kubebuilder:subresource:status
type RolloutPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RolloutPlugin `json:"items"`
}

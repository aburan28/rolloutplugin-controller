// +kubebuilder:object:generate=true
// +groupName=rolloutsplugin.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Selector Selector `json:"selector"`
}

type Selector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}

type Strategy struct {
	Type   string `json:"type"`
	Canary Canary `json:"canary"`
}

type Canary struct {
	Steps          []CanaryStep   `json:"steps"`
	TrafficRouting TrafficRouting `json:"trafficRouting"`
}

type TrafficRouting struct {
	// Istio defines the traffic routing for Istio
	// +optional
	Istio *IstioTrafficRouting `json:"istio,omitempty" protobuf:"bytes,1,opt,name=istio"`
}

type IstioTrafficRouting struct {
	// VirtualService holds the name of the VirtualService to use for routing traffic
	VirtualService IstioVirtualService `json:"virtualService" protobuf:"bytes,1,opt,name=virtualService"`
	// DestinationRule holds the name of the DestinationRule to use for routing traffic
	// +optional
	DestinationRule IstioDestinationRule `json:"destinationRule,omitempty" protobuf:"bytes,2,opt,name=destinationRule"`
	// ManagedRoutes holds the list of routes that are managed by the Rollout
	// +optional
	ManagedRoutes []string `json:"managedRoutes,omitempty" protobuf:"bytes,3,opt,name=managedRoutes"`
	// RouteWeight defines the weight of the traffic to the newRS
	// +optional
	RouteWeight *int32 `json:"routeWeight,omitempty" protobuf:"varint,4,opt,name=routeWeight"`
	// RequestHeaders defines the headers to use for routing traffic
	// +optional
	RequestHeaders []string `json:"requestHeaders,omitempty" protobuf:"bytes,5,opt,name=requestHeaders"`
	// RequestHeaders defines the headers to use for routing traffic
	// +optional
	MatchHeaders []string `json:"matchHeaders,omitempty" protobuf:"bytes,6,opt,name=matchHeaders"`
}

// IstioVirtualService holds information on the virtual service the rollout needs to modify
type IstioVirtualService struct {
	// Name holds the name of the VirtualService
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// A list of HTTP routes within VirtualService to edit. If omitted, VirtualService must have a single route of this type.
	Routes []string `json:"routes,omitempty" protobuf:"bytes,2,rep,name=routes"`
	// A list of TLS/HTTPS routes within VirtualService to edit. If omitted, VirtualService must have a single route of this type.
	TLSRoutes []TLSRoute `json:"tlsRoutes,omitempty" protobuf:"bytes,3,rep,name=tlsRoutes"`
	// A list of TCP routes within VirtualService to edit. If omitted, VirtualService must have a single route of this type.
	TCPRoutes []TCPRoute `json:"tcpRoutes,omitempty" protobuf:"bytes,4,rep,name=tcpRoutes"`
}

// TLSRoute holds the information on the virtual service's TLS/HTTPS routes that are desired to be matched for changing weights.
type TLSRoute struct {
	// Port number of the TLS Route desired to be matched in the given Istio VirtualService.
	Port int64 `json:"port,omitempty" protobuf:"bytes,1,opt,name=port"`
	// A list of all the SNI Hosts of the TLS Route desired to be matched in the given Istio VirtualService.
	SNIHosts []string `json:"sniHosts,omitempty" protobuf:"bytes,2,rep,name=sniHosts"`
}

// TCPRoute holds the information on the virtual service's TCP routes that are desired to be matched for changing weights.
type TCPRoute struct {
	// Port number of the TCP Route desired to be matched in the given Istio VirtualService.
	Port int64 `json:"port,omitempty" protobuf:"bytes,1,opt,name=port"`
}

// IstioDestinationRule is a reference to an Istio DestinationRule to modify and shape traffic
type IstioDestinationRule struct {
	// Name holds the name of the DestinationRule
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// CanarySubsetName is the subset name to modify labels with canary ReplicaSet pod template hash value
	CanarySubsetName string `json:"canarySubsetName" protobuf:"bytes,2,opt,name=canarySubsetName"`
	// StableSubsetName is the subset name to modify labels with stable ReplicaSet pod template hash value
	StableSubsetName string `json:"stableSubsetName" protobuf:"bytes,3,opt,name=stableSubsetName"`
}

type CanaryStep struct {
	// SetWeight sets what percentage of the newRS should receive
	SetWeight *int32 `json:"setWeight,omitempty" protobuf:"varint,1,opt,name=setWeight"`
	// Pause freezes the rollout by setting spec.Paused to true.
	// A Rollout will resume when spec.Paused is reset to false.
	// +optional
	Pause *RolloutPause `json:"pause,omitempty" protobuf:"bytes,2,opt,name=pause"`

	SetCanaryScale *SetCanaryScale `json:"setCanaryScale,omitempty" protobuf:"bytes,5,opt,name=setCanaryScale"`
	// SetHeaderRoute defines the route with specified header name to send 100% of traffic to the canary service
	// +optional
	SetHeaderRoute *SetHeaderRoute `json:"setHeaderRoute,omitempty" protobuf:"bytes,6,opt,name=setHeaderRoute"`
	// SetMirrorRoutes Mirrors traffic that matches rules to a particular destination
	// +optional
	SetMirrorRoute *SetMirrorRoute `json:"setMirrorRoute,omitempty" protobuf:"bytes,8,opt,name=setMirrorRoute"`
	// Plugin defines a plugin to execute for a step

}

type RolloutPause struct {
	// Duration the amount of time to wait before moving to the next step.
	// +optional
	Duration *intstr.IntOrString `json:"duration,omitempty" protobuf:"bytes,1,opt,name=duration"`
}

type RouteMatch struct {
	// Method What http methods should be mirrored
	// +optional
	Method *StringMatch `json:"method,omitempty" protobuf:"bytes,1,opt,name=method"`
	// Path What url paths should be mirrored
	// +optional
	Path *StringMatch `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`
	// Headers What request with matching headers should be mirrored
	// +optional
	Headers map[string]StringMatch `json:"headers,omitempty" protobuf:"bytes,3,opt,name=headers"`
}

type SetMirrorRoute struct {
	// Name this is the name of the route to use for the mirroring of traffic this also needs
	// to be included in the `spec.strategy.canary.trafficRouting.managedRoutes` field
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// Match Contains a list of rules that if mated will mirror the traffic to the services
	// +optional
	Match []RouteMatch `json:"match,omitempty" protobuf:"bytes,2,opt,name=match"`

	// Services The list of services to mirror the traffic to if the method, path, headers match
	//Service string `json:"service" protobuf:"bytes,3,opt,name=service"`
	// Percentage What percent of the traffic that matched the rules should be mirrored
	Percentage *int32 `json:"percentage,omitempty" protobuf:"varint,4,opt,name=percentage"`
}
type SetHeaderRoute struct {
	// Name this is the name of the route to use for the mirroring of traffic this also needs
	// to be included in the `spec.strategy.canary.trafficRouting.managedRoutes` field
	Name  string               `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	Match []HeaderRoutingMatch `json:"match,omitempty" protobuf:"bytes,2,rep,name=match"`
}

type HeaderRoutingMatch struct {
	// HeaderName the name of the request header
	HeaderName string `json:"headerName" protobuf:"bytes,1,opt,name=headerName"`
	// HeaderValue the value of the header
	HeaderValue *StringMatch `json:"headerValue" protobuf:"bytes,2,opt,name=headerValue"`
}

type StringMatch struct {
	// Exact The string must match exactly
	Exact string `json:"exact,omitempty" protobuf:"bytes,1,opt,name=exact"`
	// Prefix The string will be prefixed matched
	Prefix string `json:"prefix,omitempty" protobuf:"bytes,2,opt,name=prefix"`
	// Regex The string will be regular expression matched
	Regex string `json:"regex,omitempty" protobuf:"bytes,3,opt,name=regex"`
}

// SetCanaryScale defines how to scale the newRS without changing traffic weight
type SetCanaryScale struct {
	// Weight sets the percentage of replicas the newRS should have
	// +optional
	Weight *int32 `json:"weight,omitempty" protobuf:"varint,1,opt,name=weight"`
	// Replicas sets the number of replicas the newRS should have
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,2,opt,name=replicas"`
	// MatchTrafficWeight cancels out previously set Replicas or Weight, effectively activating SetWeight
	// +optional
	MatchTrafficWeight bool `json:"matchTrafficWeight,omitempty" protobuf:"varint,3,opt,name=matchTrafficWeight"`
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
	CurrentStepIndex   int32       `json:"currentStepIndex"`
}

type Condition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastUpdateTime     metav1.Time `json:"lastUpdateTime" protobuf:"bytes,3,opt,name=lastUpdateTime"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	Reason             Reason      `json:"reason"`
	Message            string      `json:"message"`
}

type Steps struct {
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

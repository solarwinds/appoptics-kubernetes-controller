package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Dashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TokenAndDataSpec     `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TokenAndDataSpec     `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Alert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TokenAndDataSpec     `json:"spec"`
	Status            Status `json:"status,omitempty"`
}

type TokenAndDataSpec struct {
	Namespace string `json:"namespace"`
	Data      string `json:"data"`
	Secret     string `json:"secret"`
}

type Status struct {
	LastUpdated string `json:"lastUpdated,omitempty"`
	ID          int    `json:"id,omitempty"`
	Hashes          Hashes    `json:"Hashes,omitempty"`
	UpdatedAt          int    `json:"updatedAt,omitempty"`
}

type Hashes struct {
	Spec []byte `json:spec,omitempty`
	AppOptics []byte `json:appoptics,omitempty`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Dashboard `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Service `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []Alert `json:"items"`
}
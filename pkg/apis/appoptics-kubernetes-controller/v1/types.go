package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppOpticsDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TokenAndDataSpec `json:"spec"`
	Status            Status           `json:"status,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppOpticsService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TokenAndDataSpec `json:"spec"`
	Status            Status           `json:"status,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppOpticsAlert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              TokenAndDataSpec `json:"spec"`
	Status            Status           `json:"status,omitempty"`
}

type TokenAndDataSpec struct {
	Namespace string `json:"namespace"`
	Data      string `json:"data"`
	Secret    string `json:"secret"`
}

type Status struct {
	LastUpdated string `json:"lastUpdated,omitempty"`
	ID          int    `json:"id,omitempty"`
	Hashes      Hashes `json:"Hashes,omitempty"`
	UpdatedAt   int    `json:"updatedAt,omitempty"`
}

type Hashes struct {
	Spec      []byte `json:"spec,omitempty"`
	AppOptics []byte `json:"appoptics,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppOpticsDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AppOpticsDashboard `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppOpticsServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AppOpticsService `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AppOpticsAlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []AppOpticsAlert `json:"items"`
}

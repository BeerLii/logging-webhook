package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LogInfo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              LogInfoSpec `json:"spec"`
}

type LogInfoSpec struct {
	Log []Logging `json:"log"`
}

type Logging struct {
	ConName   string   `json:"containername"`
	LogPath   string   `json:"logpath"`
	LogDetail []Detail `json:"logdetail"`
}

type Detail struct {
	FileName string `json:"filename"`
	Name     string `json:"name"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LogInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LogInfo `json:"items"`
}

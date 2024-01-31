/*
Copyright 2024.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SentinelSpec defines the desired state of Sentinel
type SentinelSpec struct {
	// The following markers will use OpenAPI v3 schema to validate the value
	// Port defines the port that will be used to init the container with the image
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	//ContainerPort int32 `json:"containerPort,omitempty"`

	// SecretName defines the name of the secret that should create
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SecretName string `json:"secretName"`

	// Data defines the key-value pair of data that should be secured
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Data map[string][]byte `json:"data,omitempty"`

	// SecretType defines the Type of the secret severity
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SecretType string `json:"secretType"`

	// ServiceAccount is optional and for the RBAC secured type
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	ServiceAccount string `json:"serviceAccount,omitempty"`

	// Role defines is optional and for the RBAC secured type
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Role string `json:"role,omitempty"`

	// RoleBinding is optional and for the RBAC secured type
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	RoleBinding string `json:"roleBinding,omitempty"`
}

// SentinelStatus defines the observed state of Sentinel
type SentinelStatus struct {
	// Represents the observations of a Sentinel's current state.
	// Sentinel.status.conditions.type are: "Available", "Progressing", and "Degraded"
	// Sentinel.status.conditions.status are one of True, False, Unknown.
	// Sentinel.status.conditions.reason the value should be a CamelCase string and producers of specific
	// condition types may define expected values and meanings for this field, and whether the values
	// are considered a guaranteed API.
	// Sentinel.status.conditions.Message is a human readable message indicating details about the transition.
	// For further information see: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// Conditions store the status conditions of the Sentinel instances
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Sentinel is the Schema for the sentinels API
type Sentinel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SentinelSpec   `json:"spec,omitempty"`
	Status SentinelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SentinelList contains a list of Sentinel
type SentinelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sentinel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sentinel{}, &SentinelList{})
}

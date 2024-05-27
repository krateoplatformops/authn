package v1alpha1

import (
	authnoauth2 "github.com/krateoplatformops/authn/apis/authn/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GithubConfigSpec struct {
	authnoauth2.ConfigSpec `json:",inline"`
	Organization           string `json:"organization"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={krateo,authn,github}

// GithubConfig is a AuthN Service Oauth2 configuration.
type GithubConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GithubConfigSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// GithubConfigList contains a list of GithubConfig
type GithubConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GithubConfig `json:"items"`
}

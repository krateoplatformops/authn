// +kubebuilder:object:generate=true
// +groupName=oauth.authn.krateo.io
// +versionName=v1alpha1
package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "oauth.authn.krateo.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// GithubConfig type metadata.
var (
	GithubConfigKind             = reflect.TypeOf(GithubConfig{}).Name()
	GithubConfigGroupKind        = schema.GroupKind{Group: Group, Kind: GithubConfigKind}.String()
	GithubConfigKindAPIVersion   = GithubConfigKind + "." + SchemeGroupVersion.String()
	GithubConfigGroupVersionKind = SchemeGroupVersion.WithKind(GithubConfigKind)
)

func init() {
	SchemeBuilder.Register(&GithubConfig{}, &GithubConfigList{})
}

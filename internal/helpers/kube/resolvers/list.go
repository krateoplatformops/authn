package resolvers

import (
	"context"
	"log"

	"github.com/krateoplatformops/authn/apis/authn/oauth2/github/v1alpha1"
	"golang.org/x/oauth2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ConfigSpec struct {
	// Kind of CRD
	Kind string `json:"kind"`

	// Name of the CRD instance with all the specs
	Name string `json:"name"`

	// AuthCodeURL: oauth2 provider authorization code URL
	AuthCodeURL string `json:"authCodeURL"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirectURL"`

	// LoginRoute path of the login handler
	LoginRoute string `json:"loginRoute"`
}

func ListGithubConfigs(dyn dynamic.Interface) ([]*ConfigSpec, error) {
	all, err := listOauth2Config(dyn, "githubconfigs")
	if err != nil {
		return nil, err
	}
	if all == nil {
		return nil, nil
	}

	for _, x := range all {
		x.Kind = "github"
		x.LoginRoute = "/github/login"
	}

	return all, nil
}

func listOauth2Config(dyn dynamic.Interface, resource string) ([]*ConfigSpec, error) {
	gvr := schema.GroupVersionResource{
		Group:    "oauth.authn.krateo.io",
		Version:  "v1alpha1",
		Resource: resource,
	}

	all, err := dyn.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	res := make([]*ConfigSpec, len(all.Items))
	for i, x := range all.Items {
		el := v1alpha1.GithubConfig{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(x.UnstructuredContent(), &el)
		if err != nil {
			log.Printf("error converting unstructured: (kind: %s, name: %s)\n", x.GetKind(), x.GetName())
			continue
		}

		// clientID, _, err := unstructured.NestedString(x.Object, "spec", "clientID")
		// if err != nil {
		// 	log.Printf("error fetching clientID for resource: (kind: %s, name: %s)\n", x.GetKind(), x.GetName())
		// 	continue
		// }

		// authURL, _, err := unstructured.NestedString(x.Object, "spec", "authURL")
		// if err != nil {
		// 	log.Printf("error fetching authURL for resource: (kind: %s, name: %s)\n", x.GetKind(), x.GetName())
		// 	continue
		// }

		// redirectURL, _, err := unstructured.NestedString(x.Object, "spec", "redirectURL")
		// if err != nil {
		// 	log.Printf("error fetching redirectURL for resource: (kind: %s, name: %s)\n", x.GetKind(), x.GetName())
		// 	continue
		// }

		oc := oauth2.Config{
			ClientID:    el.Spec.ClientID,
			RedirectURL: el.Spec.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  el.Spec.AuthURL,
				TokenURL: el.Spec.TokenURL,
			},
			Scopes: el.Spec.Scopes,
		}

		res[i] = &ConfigSpec{
			Name:        x.GetName(),
			AuthCodeURL: oc.AuthCodeURL(el.GetName()),
			RedirectURL: el.Spec.RedirectURL,
		}
	}

	return res, err
}

package resolvers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	githubv1alpha1 "github.com/krateoplatformops/authn/apis/authn/oauth2/github/v1alpha1"
	"github.com/krateoplatformops/authn/internal/helpers/kube/client"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
)

func GetGithubConfig(rc *rest.Config, name string) (*githubv1alpha1.GithubConfig, error) {
	ns, err := util.GetOperatorNamespace()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve service namespace: %w", err)
	}

	cli, err := client.New(rc, schema.GroupVersion{
		Group:   githubv1alpha1.Group,
		Version: githubv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &githubv1alpha1.GithubConfig{}
	err = cli.Get().Resource("githubconfigs").
		Namespace(ns).Name(name).
		Do(context.Background()).
		Into(res)

	return res, err
}

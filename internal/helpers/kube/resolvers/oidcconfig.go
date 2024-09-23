package resolvers

import (
	"context"
	"fmt"

	oidcv1alpha1 "github.com/krateoplatformops/authn/apis/authn/oidc/v1alpha1"
	"github.com/krateoplatformops/authn/internal/helpers/kube/client"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func OIDCConfigGet(rc *rest.Config, name string) (*oidcv1alpha1.OIDCConfig, error) {
	ns, err := util.GetOperatorNamespace()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve service namespace: %w", err)
	}

	cli, err := client.New(rc, schema.GroupVersion{
		Group:   oidcv1alpha1.Group,
		Version: oidcv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &oidcv1alpha1.OIDCConfig{}
	err = cli.Get().Resource("oidcconfigs").
		Namespace(ns).Name(name).
		Do(context.Background()).
		Into(res)

	return res, err
}

func OIDCConfigList(rc *rest.Config) (*oidcv1alpha1.OIDCConfigList, error) {
	ns, err := util.GetOperatorNamespace()
	if err != nil {
		return nil, fmt.Errorf("unable to resolve service namespace: %w", err)
	}

	cli, err := client.New(rc, schema.GroupVersion{
		Group:   oidcv1alpha1.Group,
		Version: oidcv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &oidcv1alpha1.OIDCConfigList{}
	err = cli.Get().Resource("oidcconfigs").
		Namespace(ns).
		Do(context.Background()).
		Into(res)

	return res, err
}

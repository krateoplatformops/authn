package resolvers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	ldapv1alpha1 "github.com/krateoplatformops/authn/apis/authn/ldap/v1alpha1"
	"github.com/krateoplatformops/authn/internal/helpers/kube/client"
)

func LDAPConfigGet(rc *rest.Config, name string) (*ldapv1alpha1.LDAPConfig, error) {
	cli, err := client.New(rc, schema.GroupVersion{
		Group:   ldapv1alpha1.Group,
		Version: ldapv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &ldapv1alpha1.LDAPConfig{}
	err = cli.Get().Resource("ldapconfigs").
		Name(name).
		Do(context.Background()).
		Into(res)

	return res, err
}

func LDAPConfigList(rc *rest.Config) (*ldapv1alpha1.LDAPConfigList, error) {
	cli, err := client.New(rc, schema.GroupVersion{
		Group:   ldapv1alpha1.Group,
		Version: ldapv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &ldapv1alpha1.LDAPConfigList{}
	err = cli.Get().Resource("ldapconfigs").
		Do(context.Background()).
		Into(res)

	return res, err
}

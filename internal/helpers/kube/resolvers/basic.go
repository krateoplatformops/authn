package resolvers

import (
	"context"

	basicv1alpha1 "github.com/krateoplatformops/authn/apis/authn/basic/v1alpha1"
	"github.com/krateoplatformops/authn/internal/helpers/kube/client"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func UserList(rc *rest.Config) ([]*basicv1alpha1.UserSpec, error) {
	cli, err := client.New(rc, schema.GroupVersion{
		Group:   basicv1alpha1.Group,
		Version: basicv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &basicv1alpha1.UserList{}
	err = cli.Get().Resource("users").Do(context.Background()).Into(res)
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return []*basicv1alpha1.UserSpec{}, nil
	}

	all := make([]*basicv1alpha1.UserSpec, len(res.Items))
	for i, x := range res.Items {
		all[i] = x.Spec.DeepCopy()
	}
	return all, err
}

func UserGet(rc *rest.Config, name string) (*basicv1alpha1.User, error) {
	cli, err := client.New(rc, schema.GroupVersion{
		Group:   basicv1alpha1.Group,
		Version: basicv1alpha1.Version,
	})
	if err != nil {
		return nil, err
	}

	res := &basicv1alpha1.User{}
	err = cli.Get().Resource("users").Name(name).
		Do(context.Background()).Into(res)
	return res, err
}

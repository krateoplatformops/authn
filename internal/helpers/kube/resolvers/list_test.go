package resolvers

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"k8s.io/client-go/dynamic"
)

func TestListOAuthConfigs(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	dyn, err := dynamic.NewForConfig(rc)
	if err != nil {
		t.Fatal(err)
	}

	res, err := ListOAuthConfigs(dyn)
	if err != nil {
		t.Fatal(err)
	}

	for _, el := range res {
		fmt.Printf("%+v\n", el)
	}
}

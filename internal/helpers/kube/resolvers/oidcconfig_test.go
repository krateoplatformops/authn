package resolvers

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
)

func TestOIDCConfigGet(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	res, err := OIDCConfigGet(rc, "oidc-example")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Discovery URL: %s\n", res.Spec.DiscoveryURL)
}

func TestOIDCConfigList(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	all, err := OIDCConfigList(rc)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("found: %d\n", len(all.Items))
}

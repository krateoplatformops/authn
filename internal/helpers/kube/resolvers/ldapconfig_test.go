package resolvers

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
)

func TestLDAPConfigGet(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	res, err := LDAPConfigGet(rc, "local")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("dialURL: %s\n", res.Spec.DialURL)
	fmt.Printf("baseDN: %s\n", res.Spec.BaseDN)
}

func TestLDAPConfigList(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	all, err := LDAPConfigList(rc)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("found: %d\n", len(all.Items))
}

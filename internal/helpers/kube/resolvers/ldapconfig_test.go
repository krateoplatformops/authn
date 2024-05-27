package resolvers

import (
	"fmt"
	"testing"
)

func TestLDAPConfigGet(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	res, err := LDAPConfigGet(rc, "ldap-local")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("dialURL: %s\n", res.Spec.DialURL)
	fmt.Printf("baseDN: %s\n", res.Spec.BaseDN)
}

func TestLDAPConfigList(t *testing.T) {
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

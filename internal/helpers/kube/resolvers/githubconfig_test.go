package resolvers

import (
	"fmt"
	"os"
	"testing"

	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
)

func TestGetGithubConfig(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	res, err := GetGithubConfig(rc, "github-example")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("authURL: %s\n", res.Spec.AuthURL)
	fmt.Printf("tokenURL: %s\n", res.Spec.TokenURL)
}

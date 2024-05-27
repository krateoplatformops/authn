package secrets

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/krateoplatformops/authn/apis/core"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGetSecret(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	ref := core.SecretKeySelector{
		Namespace: "krateo-system",
		Name:      "ldap-example",
		Key:       "password",
	}
	pwd, err := Get(context.TODO(), rc, &ref)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("password: %s\n", pwd)
}

func newRestConfig() (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
}

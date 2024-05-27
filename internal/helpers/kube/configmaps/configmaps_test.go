package configmaps

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGetConfigMap(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv(util.NamespaceEnvVar, "demo-system")
	caCrt, err := CACrt(context.TODO(), rc)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", caCrt)
}

func newRestConfig() (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
}

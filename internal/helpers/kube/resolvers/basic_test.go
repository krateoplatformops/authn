package resolvers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestUserGet(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	usr, err := UserGet(rc, "cyberjoker")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("username: %s\n", usr.Name)
	fmt.Printf("displayName: %s\n", usr.Spec.DisplayName)
	fmt.Printf("avatarURL: %s\n", usr.Spec.AvatarURL)
}

func TestUserList(t *testing.T) {
	os.Setenv(util.NamespaceEnvVar, "demo-system")

	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	all, err := UserList(rc)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("found: %d\n", len(all))
}

func newRestConfig() (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
}

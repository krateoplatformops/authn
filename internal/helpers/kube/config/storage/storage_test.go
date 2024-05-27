package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krateoplatformops/authn/internal/helpers/kube/util"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func TestStorage(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	os.Setenv(util.NamespaceEnvVar, "default")

	want := &AuthInfo{
		CertData: "XXX",
		KeyData:  "YYY",
		CAData:   "ZZZ",
		Server:   "AAA",
		ProxyURL: "BBB",
	}

	store := Default(rc)

	err = store.Put("test", want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("test")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(got, want); len(diff) > 0 {
		t.Fatal(diff)
	}
}

func newRestConfig() (*rest.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return clientcmd.BuildConfigFromFlags("", filepath.Join(home, ".kube", "config"))
}

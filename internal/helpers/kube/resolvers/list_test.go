package resolvers

import (
	"fmt"
	"testing"

	"k8s.io/client-go/dynamic"
)

func TestListGithubConfigs(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	dyn, err := dynamic.NewForConfig(rc)
	if err != nil {
		t.Fatal(err)
	}

	res, err := ListGithubConfigs(dyn)
	if err != nil {
		t.Fatal(err)
	}

	for _, el := range res {
		fmt.Printf("%+v\n", el)
	}
}

package resolvers

import (
	"fmt"
	"testing"
)

func TestGetGithubConfig(t *testing.T) {
	rc, err := newRestConfig()
	if err != nil {
		t.Fatal(err)
	}

	res, err := GetGithubConfig(rc, "github")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("authURL: %s\n", res.Spec.AuthURL)
	fmt.Printf("tokenURL: %s\n", res.Spec.TokenURL)
}

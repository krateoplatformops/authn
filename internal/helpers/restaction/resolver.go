package restaction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/krateoplatformops/authn/apis/core"
	xcontext "github.com/krateoplatformops/plumbing/context"
	"k8s.io/client-go/rest"
)

type RestActionContextKey string

type Response struct {
	Status map[string]interface{} `json:"status"`
}

func Resolve(ctx context.Context, rc *rest.Config, restaction *core.ObjectRef, email string, bearerToken string) (map[string]interface{}, error) {
	jsonToken, _ := json.Marshal(map[string]string{
		"token": bearerToken,
	})
	// Call the RESTAction from snowplow
	url := fmt.Sprintf(ctx.Value(RestActionContextKey("snowplowURL")).(string)+
		"/call?apiVersion=templates.krateo.io%%2Fv1&resource=restactions&name=%s&namespace=%s&extras=%s",
		restaction.Name,
		restaction.Namespace,
		jsonToken)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request for restaction call to snowplow: %w", err)
	}

	jwt, ok := xcontext.AccessToken(ctx)
	if !ok {
		return nil, fmt.Errorf("failed to retrieve jwt token for authn")
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request for restaction call to snowplow: %w", err)
	}
	responseRaw, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	var responseStruct Response
	err = json.Unmarshal(responseRaw, &responseStruct)
	if err != nil {
		return nil, fmt.Errorf("error parsing userinfo payload JSON: %v\npayload: %s", err, string(responseRaw))
	}

	return responseStruct.Status, nil
}

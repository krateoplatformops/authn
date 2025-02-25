package signup

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/krateoplatformops/snowplow/plumbing/endpoints"
	"github.com/krateoplatformops/snowplow/plumbing/kubeutil"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type SignupHandler struct {
	CAData       string
	ProxyURL     string
	ServerURL    string
	CertDuration time.Duration
	Restconfig   *rest.Config
	Namespace    string
}

func (g *SignupHandler) SignUp(user string, groups []string) (endpoints.Endpoint, error) {
	if len(g.CAData) == 0 {
		caCrt, err := kubeutil.CACrt(context.Background(), g.Restconfig)
		if err != nil {
			return endpoints.Endpoint{}, err
		}
		g.CAData = caCrt
	}

	ep, err := g.generateEndpoint(user, groups)
	if err != nil {
		return endpoints.Endpoint{}, err
	}
	ep.Username = user

	err = endpoints.Store(context.TODO(), g.Restconfig, g.Namespace, ep)
	return ep, err
}

func (g *SignupHandler) generateEndpoint(user string, groups []string) (ep endpoints.Endpoint, err error) {
	if len(g.ServerURL) == 0 {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			return ep, rest.ErrNotInCluster
		}
		g.ServerURL = "https://" + net.JoinHostPort(host, port)
	}

	cli, err := kubernetes.NewForConfig(g.Restconfig)
	if err != nil {
		return ep, err
	}

	cert, key, err := generateClientCertAndKey(cli, generateClientCertAndKeyOpts{
		userID:   mkID(fmt.Sprintf("%s@%s", user, strings.Join(groups, ","))),
		username: user,
		groups:   groups,
		duration: g.CertDuration,
	})
	if err != nil {
		return ep, err
	}

	ep.ServerURL = g.ServerURL
	ep.CertificateAuthorityData = g.CAData
	ep.ClientCertificateData = cert
	ep.ClientKeyData = key

	return
}

type generateClientCertAndKeyOpts struct {
	duration time.Duration
	userID   string
	username string
	groups   []string
}

func generateClientCertAndKey(client kubernetes.Interface, o generateClientCertAndKeyOpts) (string, string, error) {
	key, err := newPrivateKey()
	if err != nil {
		return "", "", err
	}

	req, err := newCertificateRequest(key, o.username, o.groups)
	if err != nil {
		return "", "", err
	}

	csr := newCertificateSigningRequest(req, o.duration, o.userID, o.username)

	err = createCertificateSigningRequests(client, csr)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return "", "", fmt.Errorf("creating CSR kubernetes object: %w", err)
		}

		if err := deleteCertificateSigningRequest(client, csr.Name); err != nil {
			return "", "", fmt.Errorf("deleting existing CSR kubernetes object: %w", err)
		}

		if err := createCertificateSigningRequests(client, csr); err != nil {
			return "", "", fmt.Errorf("creating CSR kubernetes object: %w", err)
		}
	}

	err = approveCertificateSigningRequest(client, csr)
	if err != nil {
		return "", "", err
	}

	err = waitForCertificate(client, csr.Name)
	if err != nil {
		return "", "", err
	}

	crt, err := certificate(client, csr.Name)
	if err != nil {
		return "", "", err
	}

	crtStr := base64.StdEncoding.EncodeToString(crt)
	keyStr := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
	return crtStr, keyStr, nil
}

func mkID(in string) string {
	hash := fnv.New64a()
	hash.Write([]byte(in))
	return strconv.FormatUint(hash.Sum64(), 16)
}

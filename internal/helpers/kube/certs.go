package kube

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"

	certv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	certificateWaitTimeout       = 5 * time.Minute
	certificateWaitPollInternval = 3 * time.Second
	resourceAnnotationKey        = "krateo.user.id"
	customSignerName             = "kubernetes.io/kube-apiserver-client"
)

func NewPrivateKey() (*rsa.PrivateKey, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generating private key: %w", err)
	}

	return key, nil
}

func NewCertificateRequest(key *rsa.PrivateKey, username string, groups []string) ([]byte, error) {
	req := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   username,
			Organization: groups,
		},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	dat, err := x509.CreateCertificateRequest(rand.Reader, &req, key)
	if err != nil {
		return nil, fmt.Errorf("creating certificate request: %w", err)
	}

	enc := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: dat})

	return enc, nil
}

func DeleteCertificateSigningRequest(client kubernetes.Interface, name string) error {
	err := client.CertificatesV1().CertificateSigningRequests().
		Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	condFunc := func(ctx context.Context) (bool, error) {
		_, err := client.CertificatesV1().CertificateSigningRequests().
			Get(ctx, name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true, nil
		}

		return false, nil
	}

	ctx := context.Background()
	err = wait.PollUntilContextTimeout(ctx, certificateWaitPollInternval, certificateWaitTimeout, true, condFunc)
	if err != nil {
		return fmt.Errorf("waiting for CSR certificate to be deleted: %w", err)
	}
	return nil
}

func CreateCertificateSigningRequests(client kubernetes.Interface, csr *certv1.CertificateSigningRequest) error {
	_, err := client.CertificatesV1().CertificateSigningRequests().
		Create(context.Background(), csr, metav1.CreateOptions{})
	return err
}

func ApproveCertificateSigningRequest(client kubernetes.Interface, csr *certv1.CertificateSigningRequest) error {
	/*
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return err
		}

		fmt.Println("rest.Config.Username: ", cfg.Username)

		client, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return err
		}
	*/
	cond := certv1.CertificateSigningRequestCondition{
		Type:           certv1.CertificateApproved,
		Status:         corev1.ConditionTrue,
		Reason:         "CertificateApproved",
		Message:        "Certificate was approved by authn-service",
		LastUpdateTime: metav1.Now(),
	}

	csr.Status.Conditions = append(csr.Status.Conditions, cond)

	ctx := context.Background()
	_, err := client.CertificatesV1().CertificateSigningRequests().
		UpdateApproval(ctx, csr.Name, csr, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("approving CertificateSigningRequest: %w", err)
	}
	return nil
}

// WaitForCertificate wait for certificate field to be generated in CSR's status.certificate field
func WaitForCertificate(client kubernetes.Interface, name string) error {
	ctx := context.Background()
	err := wait.PollUntilContextTimeout(ctx, certificateWaitPollInternval,
		certificateWaitTimeout, false, CertificateExistsFunc(client, name))
	if err != nil {
		return fmt.Errorf("waiting for CSR certificate to be generated: %w", err)
	}

	return nil
}

func CertificateExistsFunc(client kubernetes.Interface, name string) wait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		obj, err := client.CertificatesV1().CertificateSigningRequests().
			Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(obj.Status.Certificate) > 0 {
			return true, nil
		}

		return false, nil
	}
}

// Certificate get the certificate from the CertificateSigningRequests status
func Certificate(client kubernetes.Interface, name string) ([]byte, error) {
	csr, err := client.CertificatesV1().CertificateSigningRequests().
		Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("getting CSR '%s': %w", name, err)
	}

	return csr.Status.Certificate, nil
}

func NewCertificateSigningRequest(csr []byte, dur time.Duration, userID, username string) *certv1.CertificateSigningRequest {
	durationSeconds := int32(dur.Seconds())
	csrObject := &certv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: username,
			Annotations: map[string]string{
				resourceAnnotationKey: userID,
			},
		},
		Spec: certv1.CertificateSigningRequestSpec{
			Request:           csr,
			SignerName:        customSignerName,
			Usages:            []certv1.KeyUsage{certv1.UsageClientAuth},
			ExpirationSeconds: &durationSeconds,
		},
	}
	return csrObject
}

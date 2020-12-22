package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	"io/ioutil"
	"os"
	"path"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/cert"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	certResources "knative.dev/pkg/webhook/certificates/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	APIServiceName = "v1.cicdapi.tmax.io"
	ServiceName    = "cicd-webhook"

	CertDir = "/tmp/approval-api"

	K8sConfigMapName = "extension-apiserver-authentication"
	K8sConfigMapKey  = "requestheader-client-ca-file"
)

// Create and Store certificates for webhook server
// server key / server cert is stored as file in CertDir
// CA bundle is stored in ValidatingWebhookConfigurations
func createCert(ctx context.Context, client client.Client) error {
	// Make directory recursively
	if err := os.MkdirAll(CertDir, os.ModePerm); err != nil {
		return err
	}

	// Get service name and namespace
	svc := ServiceName
	ns, err := utils.Namespace()
	if err != nil {
		return err
	}

	// Create certs
	tlsKey, tlsCrt, caCrt, err := certResources.CreateCerts(ctx, svc, ns, time.Now().AddDate(1, 0, 0))
	if err != nil {
		return err
	}

	// Write certs to file
	keyPath := path.Join(CertDir, "tls.key")
	err = ioutil.WriteFile(keyPath, tlsKey, 0644)
	if err != nil {
		return err
	}

	crtPath := path.Join(CertDir, "tls.crt")
	err = ioutil.WriteFile(crtPath, tlsCrt, 0644)
	if err != nil {
		return err
	}

	// Update ApiService
	apiService := &apiregv1.APIService{}
	if err := client.Get(ctx, types.NamespacedName{Name: APIServiceName}, apiService); err != nil {
		return err
	}
	apiService.Spec.CABundle = caCrt
	if err := client.Update(ctx, apiService); err != nil {
		return err
	}

	return nil
}

func tlsConfig(ctx context.Context, client client.Client) (*tls.Config, error) {
	caPool, err := getCAPool(ctx, client)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		ClientCAs:  caPool,
		ClientAuth: tls.VerifyClientCertIfGiven,
	}, nil
}

func getCAPool(ctx context.Context, client client.Client) (*x509.CertPool, error) {
	cm := &corev1.ConfigMap{}
	if err := client.Get(ctx, types.NamespacedName{Name: K8sConfigMapName, Namespace: metav1.NamespaceSystem}, cm); err != nil {
		return nil, err
	}

	clientCA, ok := cm.Data[K8sConfigMapKey]
	if !ok {
		return nil, fmt.Errorf("no key [%s] found in configmap %s/%s", K8sConfigMapKey, metav1.NamespaceSystem, K8sConfigMapName)
	}

	certs, err := cert.ParseCertsPEM([]byte(clientCA))
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	for _, c := range certs {
		pool.AddCert(c)
	}

	return pool, nil
}

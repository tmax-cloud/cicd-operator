/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	apiregv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_createCert(t *testing.T) {
	tc := map[string]struct {
		apiService *apiregv1.APIService
		certDirRO  bool
		roFile     string

		errorOccurs  bool
		errorMessage string
	}{
		"normal": {
			apiService: &apiregv1.APIService{
				ObjectMeta: metav1.ObjectMeta{Name: APIServiceName},
			},
		},
		"mkdirErr": {
			certDirRO:    true,
			errorOccurs:  true,
			errorMessage: "approval-api",
		},
		"writeKeyErr": {
			apiService: &apiregv1.APIService{
				ObjectMeta: metav1.ObjectMeta{Name: APIServiceName},
			},
			roFile:       path.Join(certDir, "tls.key"),
			errorOccurs:  true,
			errorMessage: "tls.key",
		},
		"writeCrtErr": {
			apiService: &apiregv1.APIService{
				ObjectMeta: metav1.ObjectMeta{Name: APIServiceName},
			},
			roFile:       path.Join(certDir, "tls.crt"),
			errorOccurs:  true,
			errorMessage: "tls.crt",
		},
		"getErr": {
			errorOccurs:  true,
			errorMessage: "apiservices.apiregistration.k8s.io \"v1.cicdapi.tmax.io\" not found",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.RemoveAll(certDir))
			if c.certDirRO {
				require.NoError(t, ioutil.WriteFile(certDir, []byte(""), 0111))
			}

			defer func() {
				_ = os.RemoveAll(certDir)
				_ = os.RemoveAll(path.Dir(c.roFile))
			}()

			if c.roFile != "" {
				require.NoError(t, os.MkdirAll(path.Dir(c.roFile), os.ModePerm))
				require.NoError(t, os.MkdirAll(c.roFile, 0111))
			}

			require.NoError(t, apiregv1.AddToScheme(scheme.Scheme))
			fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			if c.apiService != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.apiService))
			}

			err := createCert(context.Background(), fakeCli)
			if c.errorOccurs {
				require.Error(t, err)
				require.Contains(t, err.Error(), c.errorMessage)
			} else {
				require.NoError(t, err)
				result := &apiregv1.APIService{}
				require.NoError(t, fakeCli.Get(context.Background(), types.NamespacedName{Name: APIServiceName}, result))
				p, _ := pem.Decode(result.Spec.CABundle)
				require.Equal(t, "CERTIFICATE", p.Type)
				cert, err := x509.ParseCertificate(p.Bytes)
				require.NoError(t, err)
				require.Equal(t, []string{"knative.dev"}, cert.Issuer.Organization)
				require.Equal(t, fmt.Sprintf("cicd-webhook.%s.svc", utils.Namespace()), cert.Issuer.CommonName)
				require.Equal(t, []string{"cicd-webhook", fmt.Sprintf("cicd-webhook.%s", utils.Namespace()), fmt.Sprintf("cicd-webhook.%s.svc", utils.Namespace()), fmt.Sprintf("cicd-webhook.%s.svc.cluster.local", utils.Namespace())}, cert.DNSNames)
			}
		})
	}
}

func Test_tlsConfig(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "extension-apiserver-authentication", Namespace: "kube-system"},
			Data: map[string]string{
				"requestheader-client-ca-file": `-----BEGIN CERTIFICATE-----
MIIBaTCCAQ6gAwIBAgIBADAKBggqhkjOPQQDAjArMSkwJwYDVQQDDCBrM3MtcmVx
dWVzdC1oZWFkZXItY2FAMTU5NTU0OTQ0NjAeFw0yMDA3MjQwMDEwNDZaFw0zMDA3
MjIwMDEwNDZaMCsxKTAnBgNVBAMMIGszcy1yZXF1ZXN0LWhlYWRlci1jYUAxNTk1
NTQ5NDQ2MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE85FjW9alEGKy8Jcavk2b
+hvgPV6XxgXnuz0G9RMxLsKu9SXVnaMRc2L9nXTnYOuz5b2FlnTdCWp7WTt35YVQ
VKMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0E
AwIDSQAwRgIhALCUqk9KPgxhXs+ka5oBnMVgP/xDd33WooGChkXCdLXXAiEA9YQX
rcFz1g2uGUgBe3mBBDID0wosv/64zWA1x4uuwuM=
-----END CERTIFICATE-----`,
			},
		}
		fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(cm).Build()
		cfg, err := tlsConfig(context.Background(), fakeCli)
		require.NoError(t, err)
		require.Equal(t, [][]uint8{{0x30, 0x2b, 0x31, 0x29, 0x30, 0x27, 0x6, 0x3, 0x55, 0x4, 0x3, 0xc, 0x20, 0x6b, 0x33, 0x73, 0x2d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2d, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2d, 0x63, 0x61, 0x40, 0x31, 0x35, 0x39, 0x35, 0x35, 0x34, 0x39, 0x34, 0x34, 0x36}}, cfg.ClientCAs.Subjects())
		require.Equal(t, tls.VerifyClientCertIfGiven, cfg.ClientAuth)
	})

	t.Run("caPoolErr", func(t *testing.T) {
		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "extension-apiserver-authentication", Namespace: "kube-system"},
			Data: map[string]string{
				"requestheader-client-ca-file": "",
			},
		}
		fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(cm).Build()
		_, err := tlsConfig(context.Background(), fakeCli)
		require.Error(t, err)
		require.Equal(t, "data does not contain any valid RSA or ECDSA certificates", err.Error())
	})
}

func Test_getCAPool(t *testing.T) {
	tc := map[string]struct {
		cm *corev1.ConfigMap

		errorOccurs      bool
		errorMessage     string
		expectedSubjects [][]byte
	}{
		"normal": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "extension-apiserver-authentication", Namespace: "kube-system"},
				Data: map[string]string{
					"requestheader-client-ca-file": `-----BEGIN CERTIFICATE-----
MIIBaTCCAQ6gAwIBAgIBADAKBggqhkjOPQQDAjArMSkwJwYDVQQDDCBrM3MtcmVx
dWVzdC1oZWFkZXItY2FAMTU5NTU0OTQ0NjAeFw0yMDA3MjQwMDEwNDZaFw0zMDA3
MjIwMDEwNDZaMCsxKTAnBgNVBAMMIGszcy1yZXF1ZXN0LWhlYWRlci1jYUAxNTk1
NTQ5NDQ2MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE85FjW9alEGKy8Jcavk2b
+hvgPV6XxgXnuz0G9RMxLsKu9SXVnaMRc2L9nXTnYOuz5b2FlnTdCWp7WTt35YVQ
VKMjMCEwDgYDVR0PAQH/BAQDAgKkMA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0E
AwIDSQAwRgIhALCUqk9KPgxhXs+ka5oBnMVgP/xDd33WooGChkXCdLXXAiEA9YQX
rcFz1g2uGUgBe3mBBDID0wosv/64zWA1x4uuwuM=
-----END CERTIFICATE-----`,
				},
			},
			expectedSubjects: [][]uint8{{0x30, 0x2b, 0x31, 0x29, 0x30, 0x27, 0x6, 0x3, 0x55, 0x4, 0x3, 0xc, 0x20, 0x6b, 0x33, 0x73, 0x2d, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2d, 0x68, 0x65, 0x61, 0x64, 0x65, 0x72, 0x2d, 0x63, 0x61, 0x40, 0x31, 0x35, 0x39, 0x35, 0x35, 0x34, 0x39, 0x34, 0x34, 0x36}},
		},
		"cmGetErr": {
			errorOccurs:  true,
			errorMessage: "configmaps \"extension-apiserver-authentication\" not found",
		},
		"noCAData": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "extension-apiserver-authentication", Namespace: "kube-system"},
				Data:       map[string]string{},
			},
			errorOccurs:  true,
			errorMessage: "no key [requestheader-client-ca-file] found in configmap kube-system/extension-apiserver-authentication",
		},
		"parseCertErr": {
			cm: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "extension-apiserver-authentication", Namespace: "kube-system"},
				Data: map[string]string{
					"requestheader-client-ca-file": "",
				},
			},
			errorOccurs:  true,
			errorMessage: "data does not contain any valid RSA or ECDSA certificates",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			if c.cm != nil {
				require.NoError(t, fakeCli.Create(context.Background(), c.cm))
			}

			pool, err := getCAPool(context.Background(), fakeCli)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedSubjects, pool.Subjects())
			}
		})
	}
}

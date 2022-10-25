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

package mail

import (
	"net"
	"net/textproto"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestNewSender(t *testing.T) {
	fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	s := NewSender(fakeCli)
	require.Equal(t, &sender{cli: fakeCli}, s)
}

type testEmailStruct struct {
	from   string
	to     []string
	header map[string]string
	data   []string
}

var testEmailResult testEmailStruct

func Test_sender_Send(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() {
		_ = l.Close()
	}()

	tc := map[string]struct {
		disabled bool
		users    []string
		title    string
		content  string
		isHTML   bool
		secret   string

		errorOccurs   bool
		errorMessage  string
		expectedEmail testEmailStruct
	}{
		"normal": {
			users:   []string{"test@tmax.co.kr", "test2@tmax.co.kr"},
			title:   "test email!!!!!!",
			content: "test test test test content",
			isHTML:  false,
			secret:  "smtp-auth",

			expectedEmail: testEmailStruct{
				from: "FROM:<admin@tmax.co.kr>",
				to:   []string{"TO:<test@tmax.co.kr>", "TO:<test2@tmax.co.kr>"},
				header: map[string]string{
					"From":         "admin@tmax.co.kr",
					"To":           "<test@tmax.co.kr>, <test2@tmax.co.kr>",
					"Content-Type": "text/plain; charset=UTF-8",
					"MIME-Version": "1.0",
					"Subject":      "test email!!!!!!",
				},
				data: []string{"test test test test content"},
			},
		},
		"emptyUser": {},
		"errGetServer": {
			users:   []string{"test@tmax.co.kr", "test2@tmax.co.kr"},
			title:   "test email!!!!!!",
			content: "test test test test content",
			isHTML:  false,
			secret:  "smtp-authhhhhhhhhhhh",

			errorOccurs:  true,
			errorMessage: "secrets \"smtp-authhhhhhhhhhhh\" not found",
		},
		"html": {
			users:   []string{"test@tmax.co.kr", "test2@tmax.co.kr"},
			title:   "test email!!!!!!",
			content: "test test test test content",
			isHTML:  true,
			secret:  "smtp-auth",

			expectedEmail: testEmailStruct{
				from: "FROM:<admin@tmax.co.kr>",
				to:   []string{"TO:<test@tmax.co.kr>", "TO:<test2@tmax.co.kr>"},
				header: map[string]string{
					"From":         "admin@tmax.co.kr",
					"To":           "<test@tmax.co.kr>, <test2@tmax.co.kr>",
					"Content-Type": "text/html; charset=UTF-8",
					"MIME-Version": "1.0",
					"Subject":      "test email!!!!!!",
				},
				data: []string{"test test test test content"},
			},
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.EnableMail = !c.disabled
			configs.SMTPHost = l.Addr().String()
			configs.SMTPUserSecret = c.secret

			testEmailResult = testEmailStruct{}
			exitCh := make(chan struct{}, 1)
			go mockSMTPServer(l, t, exitCh)

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "smtp-auth",
					Namespace: utils.Namespace(),
				},
				Type: corev1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					corev1.BasicAuthUsernameKey: []byte("admin@tmax.co.kr"),
					corev1.BasicAuthPasswordKey: []byte("admin"),
				},
			}
			fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(secret).Build()
			sender := &sender{cli: fakeCli}
			err = sender.Send(c.users, c.title, c.content, c.isHTML)
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)

				require.Equal(t, c.expectedEmail, testEmailResult)
			}
		})
	}
}

func Test_sender_auth(t *testing.T) {
	s := &sender{}
	srv := &smtpInfo{
		host:     "smtp.test.com:2323",
		user:     "test-user",
		password: "test-pw",
	}
	a := s.auth(srv)
	t.Log(a)
}

func Test_sender_getServerInfo(t *testing.T) {
	tc := map[string]struct {
		smtpSecret string

		errorOccurs      bool
		errorMessage     string
		expectedSmtpInfo smtpInfo
	}{
		"normal": {
			smtpSecret: "smtp-auth",
			expectedSmtpInfo: smtpInfo{
				host:     "host.smtp.test:2323",
				user:     "admin@tmax.co.kr",
				password: "admin",
			},
		},
		"getErr": {
			smtpSecret:   "asdasd",
			errorOccurs:  true,
			errorMessage: "secrets \"asdasd\" not found",
		},
		"typeErr": {
			smtpSecret:   "smtp-auth-wrong-type",
			errorOccurs:  true,
			errorMessage: "secret smtp-auth-wrong-type should be in type kubernetes.io/basic-auth (is kubernetes.io/dockercfg now), and have both keys username, password",
		},
	}

	for name, c := range tc {
		t.Run(name, func(t *testing.T) {
			configs.SMTPHost = "host.smtp.test:2323"
			configs.SMTPUserSecret = c.smtpSecret

			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "smtp-auth",
					Namespace: utils.Namespace(),
				},
				Type: corev1.SecretTypeBasicAuth,
				Data: map[string][]byte{
					corev1.BasicAuthUsernameKey: []byte("admin@tmax.co.kr"),
					corev1.BasicAuthPasswordKey: []byte("admin"),
				},
			}

			secretWrong := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "smtp-auth-wrong-type",
					Namespace: utils.Namespace(),
				},
				Type: corev1.SecretTypeDockercfg,
				Data: map[string][]byte{
					corev1.BasicAuthUsernameKey: []byte("admin@tmax.co.kr"),
					corev1.BasicAuthPasswordKey: []byte("admin"),
				},
			}

			fakeCli := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(secret, secretWrong).Build()
			s := &sender{cli: fakeCli}

			info, err := s.getServerInfo()
			if c.errorOccurs {
				require.Error(t, err)
				require.Equal(t, c.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, c.expectedSmtpInfo, *info)
			}
		})
	}
}

func mockSMTPServer(l net.Listener, t *testing.T, exitCh chan struct{}) {
	conn, err := l.Accept()
	if err != nil {
		for {
			select {
			case <-exitCh:
				return
			default:
				return
			}
		}
	}
	defer func() {
		_ = conn.Close()
	}()

	tc := textproto.NewConn(conn)
	require.NoError(t, tc.PrintfLine("220 hello world"))

	msg, err := tc.ReadLine()
	require.NoError(t, err)
	require.Equal(t, "EHLO localhost", msg)

	require.NoError(t, tc.PrintfLine("Hello localhost"))
	require.NoError(t, tc.PrintfLine("250 AUTH LOGIN PLAIN"))

	msg, err = tc.ReadLine()
	require.NoError(t, err)
	require.Equal(t, "HELO localhost", msg)

	require.NoError(t, tc.PrintfLine("250 AUTH LOGIN PLAIN"))

	isData := false
	isHeader := false
	for {
		id := tc.Next()

		msg, err = tc.ReadLine()
		require.NoError(t, err)
		t.Logf("REQ: %s\n", msg)

		if isData {
			handleData(t, tc, id, msg, &isData, &isHeader)
			continue
		}

		tc.StartResponse(id)
		cmd := msg[:4]
		var resp string
		switch cmd {
		case "MAIL":
			testEmailResult.from = msg[5:]
		case "RCPT":
			testEmailResult.to = append(testEmailResult.to, msg[5:])
			resp = "250 OK"
		case "DATA":
			isData = true
			isHeader = true
			testEmailResult.header = map[string]string{}
			resp = "354 Go ahead"
		case "QUIT":
			resp = "221 Good bye"
		}

		respond(t, tc, resp)
		tc.EndResponse(id)
		if cmd == "QUIT" {
			break
		}
	}
}

func handleData(t *testing.T, tc *textproto.Conn, id uint, msg string, isData, isHeader *bool) {
	tc.StartResponse(id)
	switch msg {
	case ".":
		respond(t, tc, "250 OK")
		*isData = false
	case "":
		if *isHeader {
			*isHeader = false
		} else {
			testEmailResult.data = append(testEmailResult.data, msg)
		}
	default:
		if *isHeader {
			tok := strings.Split(msg, ": ")
			testEmailResult.header[tok[0]] = tok[1]
		} else {
			testEmailResult.data = append(testEmailResult.data, msg)
		}
	}
	tc.EndResponse(id)
}

func respond(t *testing.T, tc *textproto.Conn, msg string) {
	if msg != "" {
		t.Logf("---> %s\n", msg)
		_ = tc.PrintfLine(msg)
	}
}

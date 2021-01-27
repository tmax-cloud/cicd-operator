package mail

import (
	"context"
	"fmt"
	"github.com/tmax-cloud/cicd-operator/internal/configs"
	"github.com/tmax-cloud/cicd-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/smtp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// Send actually sends email using SMTP call
func Send(to []string, subject string, content string, isHTML bool, c client.Client) error {
	if !configs.EnableMail {
		return fmt.Errorf("email is disabled")
	}

	if len(to) < 1 {
		return nil
	}

	server, err := serverInfo(c)
	if err != nil {
		return err
	}

	auth, err := auth(server)
	if err != nil {
		return err
	}

	toStr := ""
	for i, t := range to {
		if i != 0 {
			toStr += ", "
		}
		toStr += "<" + t + ">"
	}

	cType := "text/plain"
	if isHTML {
		cType = "text/html"
	}

	from := server.user

	header := make(map[string]string)
	header["From"] = from
	header["To"] = toStr
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = fmt.Sprintf("%s; charset=\"utf-8\"", cType)

	msg := ""
	for k, v := range header {
		msg += k + ": " + v + "\r\n"
	}
	msg += "\r\n" + content

	return smtp.SendMail(server.host, auth, from, to, []byte(msg))
}

type smtpInfo struct {
	host     string
	user     string
	password string
}

func serverInfo(c client.Client) (*smtpInfo, error) {
	ns, err := utils.Namespace()
	if err != nil {
		return nil, err
	}

	secret := &corev1.Secret{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: configs.SMTPUserSecret, Namespace: ns}, secret); err != nil {
		return nil, err
	}

	username, nameExist := secret.Data[corev1.BasicAuthUsernameKey]
	password, pwExist := secret.Data[corev1.BasicAuthPasswordKey]

	if secret.Type != corev1.SecretTypeBasicAuth || !nameExist || !pwExist {
		return nil, fmt.Errorf("secret %s should be in type %s (is %s now), and have both keys %s, %s", configs.SMTPUserSecret, corev1.SecretTypeBasicAuth, secret.Type, corev1.BasicAuthUsernameKey, corev1.BasicAuthPasswordKey)
	}

	info := &smtpInfo{
		host:     configs.SMTPHost,
		user:     string(username),
		password: string(password),
	}
	return info, nil
}

func auth(server *smtpInfo) (auth smtp.Auth, err error) {
	hosts := strings.Split(server.host, ":")
	return smtp.PlainAuth("", server.user, server.password, hosts[0]), nil
}

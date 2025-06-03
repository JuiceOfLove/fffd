package mail

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type MailService struct {
	dialer *gomail.Dialer
}

func NewMailService() *MailService {
	host := os.Getenv("SMTP_HOST")
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	return &MailService{
		dialer: gomail.NewDialer(host, port, user, pass),
	}
}

func (m *MailService) SendActivationMail(to, activationLink string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTP_USER"))
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Активация аккаунта на "+os.Getenv("CLIENT_URL"))
	message.SetBody("text/html", `
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px; background-color: #f5f5f5;">
			<h2 style="color: #333; text-align: center;">Активация аккаунта на `+os.Getenv("CLIENT_URL")+`</h2>
			<p>Здравствуйте,</p>
			<p>Для завершения регистрации, перейдите по ссылке ниже для активации аккаунта:</p>
			<p style="text-align: center;"><a href="`+activationLink+`" style="display: inline-block; padding: 10px 20px; background-color: #007bff; color: #fff; text-decoration: none; border-radius: 5px;">Активировать аккаунт</a></p>
			<p>С уважением, команда FP.</p>
		</div>
	`)
	return m.dialer.DialAndSend(message)
}

func (m *MailService) SendFamilyInviteMail(to, inviteLink string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", os.Getenv("SMTP_USER"))
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Приглашение присоединиться к семье на FP")
	message.SetBody("text/html", `
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: auto; padding: 20px; border: 1px solid #ddd; border-radius: 8px; background-color: #f5f5f5;">
			<h2 style="color: #333; text-align: center;">Приглашение в семью</h2>
			<p>Здравствуйте,</p>
			<p>Вы приглашены присоединиться к семье на FP. Для подтверждения приглашения перейдите по ссылке ниже:</p>
			<p style="text-align: center;"><a href="`+inviteLink+`" style="display: inline-block; padding: 10px 20px; background-color: #28a745; color: #fff; text-decoration: none; border-radius: 5px;">Принять приглашение</a></p>
			<p>Если вы не регистрировались на FP, сначала зарегистрируйтесь, затем используйте приглашение.</p>
			<p>С уважением, команда FP.</p>
		</div>
	`)
	return m.dialer.DialAndSend(message)
}
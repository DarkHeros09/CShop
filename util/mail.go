package util

import (
	"fmt"
	"net/smtp"
)

func EmailSend(email string, newPassword string, hostPassword string) error {
	from := "mnbenghuzzi@gmail.com"
	password := hostPassword

	// Receiver email address.
	to := []string{email}

	// smtp server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Message.

	subject := "Subject: Golang Account Recovery\n"

	mainMessage := fmt.Sprintf("<body>Your password  verification code is <h2 style=\"text-align:center;\"><span style=\"font-size:40px;border:2px solid black;padding:10px\">%v</span></h2> \n</body>", newPassword)

	body := mainMessage
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message := []byte(subject + mime + body)

	// Authentication.
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}
	fmt.Println("Email Sent Successfully!")

	return nil

}

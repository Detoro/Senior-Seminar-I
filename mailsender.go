package main

import (
	"net/smtp"
    "log"
)

func SendEmail(to string, subject string, body string) {
    from := "myvcstester@gmail.com"
    password := "NotARealPassword"
    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    auth := smtp.PlainAuth("", from, password, smtpHost)
    msg := []byte("To: " + to + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "\r\n" +
        body + "\r\n")

    err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, msg)
    if err != nil {
        log.Fatal(err)
    }
}
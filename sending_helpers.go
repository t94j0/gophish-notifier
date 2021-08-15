package main

import (
	"fmt"
	"net/smtp"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/spf13/viper"
)

func sendSlackAttachment(attachment slack.Attachment) error {
	payload := slack.Payload{
		Username:    viper.GetString("bot_username"),
		Channel:     viper.GetString("bot_channel"),
		IconEmoji:   viper.GetString("bot_emoji"),
		Attachments: []slack.Attachment{attachment},
	}
	if errs := slack.Send(viper.GetString("slack_webhook"), "", payload); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func sendEmail(subject, body string) error {
	from := viper.GetString("email_sender")
	pass := viper.GetString("email_sender_password")
	to := viper.GetString("email_recipient")
	hostAddr := viper.GetString("email_host_addr")
	host := viper.GetString("email_host")

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)

	plainAuth := smtp.PlainAuth("", from, pass, host)
	if err := smtp.SendMail(hostAddr, plainAuth, from, []string{to}, []byte(msg)); err != nil {
		return err
	}
	return nil
}

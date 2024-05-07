package main

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
)

type ghostwriter_oplog_entry struct {
	Oplog           int
	StartDate       string
	EndDate         string
	SourceIp        string
	DestIp          string
	Tool            string
	UserContext     string
	Command         string
	Description     string
	Output          string
	Comments        string
	OperatorName    string
	EntryIdentifier string
	ExtraFields     string
}

func sendSlackAttachment(attachment slack.Attachment) error {
	payload := slack.Payload{
		Username:    viper.GetString("slack.bot_username"),
		Channel:     viper.GetString("slack.bot_channel"),
		IconEmoji:   viper.GetString("slack.bot_emoji"),
		Attachments: []slack.Attachment{attachment},
	}
	if errs := slack.Send(viper.GetString("slack.webhook"), "", payload); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func sendEmail(subject, body string) error {
	from := viper.GetString("email.sender")
	pass := viper.GetString("email.sender_password")
	to := viper.GetString("email.recipient")
	hostAddr := viper.GetString("email.host_addr")
	host := viper.GetString("email.host")

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s", from, to, subject, body)

	plainAuth := smtp.PlainAuth("", from, pass, host)
	if err := smtp.SendMail(hostAddr, plainAuth, from, []string{to}, []byte(msg)); err != nil {
		return err
	}
	return nil
}

func sendGraphql(data ghostwriter_oplog_entry) error {
	url := viper.GetString("ghostwriter.graphql_endpoint")
	query := viper.GetString("graphql_default_query")
	oplog_id := viper.GetInt("ghostwriter.oplog_id")
	client := graphql.NewClient(url)

	req := graphql.NewRequest(query)
	req.Var("oplog", oplog_id)
	req.Var("sourceIp", data.SourceIp)
	req.Var("tool", "gophish")
	req.Var("userContext", data.UserContext)
	req.Var("description", data.Description)
	req.Var("output", data.Output)
	req.Var("comments", data.Comments)

	ctx := context.Background()
	var respData map[string]interface{}
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil
}

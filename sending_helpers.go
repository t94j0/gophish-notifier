package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/machinebox/graphql"
	"github.com/spf13/viper"
)

type ghostwriterOplogEntry struct {
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

func sendGraphql(data ghostwriterOplogEntry) error {
	url := viper.GetString("ghostwriter.graphql_endpoint")
	apiKey := viper.GetString("ghostwriter.api_key")
	ignoreSelfSigned := viper.GetBool("ghostwriter.ignore_self_signed_certificate")
	query := viper.GetString("graphql_default_query")
	oplogId := viper.GetInt("ghostwriter.oplog_id")

	var client *graphql.Client
	if ignoreSelfSigned {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient := &http.Client{Transport: tr}
		client = graphql.NewClient(url, graphql.WithHTTPClient(httpClient))
	} else {
		client = graphql.NewClient(url)
	}

	req := graphql.NewRequest(query)
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Var("oplog", oplogId)
	if data.SourceIp != "" {
		req.Var("sourceIp", data.SourceIp)
	}
	req.Var("tool", "gophish")
	if data.UserContext != "" {
		req.Var("userContext", data.UserContext)
	}
	if data.Description != "" {
		req.Var("description", data.Description)
	}
	if data.Output != "" {
		req.Var("output", data.Output)
	}
	if data.Comments != "" {
		req.Var("comments", data.Comments)
	}
	req.Var("extraFields", "")

	ctx := context.Background()
	var respData map[string]interface{}
	if err := client.Run(ctx, req, &respData); err != nil {
		return err
	}

	return nil
}

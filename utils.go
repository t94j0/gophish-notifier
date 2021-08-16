package main

import (
	"fmt"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/spf13/viper"
)

func addCampaignButton(attachment slack.Attachment, campaignID uint) slack.Attachment {
	gophishBase := viper.GetString("base_url")
	if gophishBase != "" {
		attachment.AddAction(slack.Action{
			Type:  "button",
			Text:  "Go to Campaign",
			Url:   fmt.Sprintf("%s/campaigns/%d", gophishBase, campaignID),
			Style: "primary",
		})
	}
	return attachment
}

func slackFormatIP(ip string) string {
	base := viper.GetString("ip_query_base")
	return fmt.Sprintf("<%s%s|%s>", base, ip, ip)
}

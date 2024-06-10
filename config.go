package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var defaultErrorTemplate = `Email ID - {{ .ID }}`

var defaultSentTemplate = `Email ID - {{ .ID }}`

var defaultClickedTemplate = `Email ID - {{ .ID }}
Email Address -  {{ .Email }}
IP Address - {{ .Address }}
User Agent - {{ .UserAgent }}`

var defaultSubmittedCredentailsTemplate = `Email ID - {{ .ID }}
Email Address - {{ .Email }}
IP Address - {{ .Address }}
User Agent - {{ .UserAgent }}
Username - {{ .Username }}
Password - {{ .Password }}`

var defaultEmailOpenedTemplate = `Email ID - {{ .ID }}
Email Address - {{ .Email }}
IP Address - {{ .Address }}
User Agent - {{ .UserAgent }}`

var defaultReportedTemplate = `Email ID - {{ .ID }}`

var defaultgraphqlTemplate = `mutation InsertGophishLog ($oplog: bigint!, $sourceIp: String, $tool: String,	$userContext: String, $description: String, $output: String, $comments: String, $extraFields: jsonb!) {
	insert_oplogEntry(objects: {oplog: $oplog, sourceIp: $sourceIp, tool: $tool, userContext: $userContext, description: $description, comments: $comments, output: $output, extraFields: $extraFields}) {
	  returning {
		id
	  }
	}
  }`

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/gophish_notifier")
	viper.AddConfigPath(".")
	setDefaults()
	if err := viper.ReadInConfig(); err != nil {
		log.Error(err)
	}
	log.Infof("Using config file: %s", viper.ConfigFileUsed())
	validateConfig()
	setLogLevel()
}

func setDefaults() {
	viper.SetDefault("log_level", "info")
	viper.SetDefault("slack.bot_username", "PhishBot")
	viper.SetDefault("slack.bot_emoji", ":blowfish:")
	viper.SetDefault("slack.disable_credentials", false)
	viper.SetDefault("listen_host", "0.0.0.0")
	viper.SetDefault("ip_query_base", "https://whatismyipaddress.com/ip/")
	viper.SetDefault("listen_port", "9999")
	viper.SetDefault("webhook_path", "/webhook")
	viper.SetDefault("email_error_sending_template", defaultErrorTemplate)
	viper.SetDefault("email_sent_template", defaultSentTemplate)
	viper.SetDefault("email_send_click_template", defaultClickedTemplate)
	viper.SetDefault("email_submitted_credentials_template", defaultSubmittedCredentailsTemplate)
	viper.SetDefault("email_default_email_open_template", defaultEmailOpenedTemplate)
	viper.SetDefault("email_reported_template", defaultReportedTemplate)
	viper.SetDefault("graphql_default_query", defaultgraphqlTemplate)
	viper.SetDefault("profiles", []string{"slack"})
	viper.SetDefault("events", []string{"email_opened", "clicked_link", "submitted_data"})
}

func setLogLevel() {
	level, err := log.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		log.Fatal("log level must be a valid level: panic, fatal, error, warning, info, debug, trace")
	}
	log.SetLevel(level)
}

func validateConfig() {
	checkKeysExist := func(keys ...string) {
		for _, key := range keys {
			if !viper.IsSet(key) {
				log.Fatal("Config value is not set: ", key)
			}
		}
	}

	globalConfigs := []string{"secret", "profiles", "events"}
	checkKeysExist(globalConfigs...)

	profiles := viper.GetStringSlice("profiles")
	for _, profile := range profiles {
		if profile == "slack" {
			slackConfigs := []string{"slack.webhook", "slack.bot_channel"}
			checkKeysExist(slackConfigs...)
			log.Infof("Using Slack sending profile. Will send messages to %s", viper.GetString("slack.bot_channel"))
			continue
		}
		if profile == "email" {
			emailConfigs := []string{"email.sender", "email.sender_password", "email.recipient", "email.host", "email.host_addr"}
			checkKeysExist(emailConfigs...)
			log.Infof("Using Email sending profile. Will send emails from %s to %s",
				viper.GetString("email.sender"),
				viper.GetString("email.recipient"))
			continue
		}
		if profile == "ghostwriter" {
			ghostwriterConfigs := []string{"ghostwriter.graphql_endpoint", "ghostwriter.api_key", "ghostwriter.oplog_id"}
			checkKeysExist(ghostwriterConfigs...)
			log.Infof("Using Ghostwriter sending profile. Will send messages to %s", viper.GetString("ghostwriter.graphql_endpoint"))
			continue
		}
		log.Fatalf("Profile \"%s\" does not exist", profile)
	}
}

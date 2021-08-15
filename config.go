package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

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

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/gophish_slack")
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
	viper.SetDefault("bot_username", "PhishBot")
	viper.SetDefault("bot_emoji", ":blowfish:")
	viper.SetDefault("listen_host", "0.0.0.0")
	viper.SetDefault("listen_port", "9999")
	viper.SetDefault("webhook_path", "/webhook")
	viper.SetDefault("email_send_click_template", defaultClickedTemplate)
	viper.SetDefault("email_submitted_credentials_template", defaultSubmittedCredentailsTemplate)
	viper.SetDefault("email_default_email_open_template", defaultEmailOpenedTemplate)
	viper.SetDefault("profiles", []string{"slack"})
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

	globalConfigs := []string{"bot_channel", "slack_webhook", "secret", "profiles"}
	checkKeysExist(globalConfigs...)

	profiles := viper.GetStringSlice("profiles")
	for _, profile := range profiles {
		if profile == "slack" {
			slackConfigs := []string{"slack_webhook"}
			checkKeysExist(slackConfigs...)
			log.Infof("Using Slack sending profile. Will send messages to %s", viper.GetString("bot_channel"))
			continue
		}
		if profile == "email" {
			emailConfigs := []string{"email_sender", "email_sender_password", "email_recipient", "email_host", "email_host_addr"}
			checkKeysExist(emailConfigs...)
			log.Infof("Using Email sending profile. Will send emails from %s to %s",
				viper.GetString("email_sender"),
				viper.GetString("email_recipient"))
			continue
		}
		log.Fatalf("Profile \"%s\" does not exist", profile)
	}
}

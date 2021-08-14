package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"github.com/ashwanthkumar/slack-go-webhook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	errNoSignature      = errors.New("header X-Gophish-Signature not provided")
	errInvalidSignature = errors.New("invalid signature")
)

// Can't be const because need reference to variable for Slack webhook title
var (
	ClickedLink   string = "Clicked Link"
	SubmittedData string = "Submitted Data"
	EmailOpened   string = "Email Opened"
)

type SlackSender interface {
	Send(WebhookResponse) error
}

type WebhookResponse struct {
	Success    string `json:"success"`
	CampaignID uint   `json:"campaign_id"`
	Message    string `json:"message"`
	Details    string `json:"details"`
	Email      string `json:"email"`
}

func sendSlack(attachment slack.Attachment) error {
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

type WebhookResponseSubmittedDetails struct {
	Payload struct {
		ID       []string `json:"id"`
		Password []string `json:"password"`
		Username []string `json:"username"`
	} `json:"payload"`
	Browser struct {
		Address   string `json:"address"`
		UserAgent string `json:"user-agent"`
	} `json:"browser"`
}

func NewWebhookResponseSubmittedDetails(detailsRaw []byte) (WebhookResponseSubmittedDetails, error) {
	var details WebhookResponseSubmittedDetails
	if err := json.Unmarshal(detailsRaw, &details); err != nil {
		return WebhookResponseSubmittedDetails{}, err
	}
	return details, nil
}

func (w WebhookResponseSubmittedDetails) Send(response WebhookResponse) error {
	red := "#f05b4f"
	attachment := slack.Attachment{Title: &SubmittedData, Color: &red}
	attachment.AddField(slack.Field{Title: "ID", Value: w.Payload.ID[0]})
	attachment.AddField(slack.Field{Title: "Email", Value: response.Email})
	attachment.AddField(slack.Field{Title: "Address", Value: w.Browser.Address})
	attachment.AddField(slack.Field{Title: "User Agent", Value: w.Browser.UserAgent})
	attachment.AddField(slack.Field{Title: "Username", Value: w.Payload.Username[0]})
	attachment.AddField(slack.Field{Title: "Password", Value: w.Payload.Password[0]})
	return sendSlack(attachment)
}

type WebhookResponseClickDetails struct {
	Payload struct {
		ID []string `json:"id"`
	} `json:"payload"`
	Browser struct {
		Address   string `json:"address"`
		UserAgent string `json:"user-agent"`
	} `json:"browser"`
}

func NewWebhookResponseClickDetails(detailsRaw []byte) (WebhookResponseClickDetails, error) {
	var details WebhookResponseClickDetails
	if err := json.Unmarshal(detailsRaw, &details); err != nil {
		return WebhookResponseClickDetails{}, err
	}
	return details, nil
}

func (w WebhookResponseClickDetails) Send(response WebhookResponse) error {
	orange := "#ffa500"
	attachment := slack.Attachment{Title: &ClickedLink, Color: &orange}
	attachment.AddField(slack.Field{Title: "ID", Value: w.Payload.ID[0]})
	attachment.AddField(slack.Field{Title: "Email", Value: response.Email})
	attachment.AddField(slack.Field{Title: "Address", Value: w.Browser.Address})
	attachment.AddField(slack.Field{Title: "User Agent", Value: w.Browser.UserAgent})
	return sendSlack(attachment)
}

type WebhookResponseOpenedDetails struct {
	Payload struct {
		ID []string `json:"id"`
	} `json:"payload"`
	Browser struct {
		Address   string `json:"address"`
		UserAgent string `json:"user-agent"`
	} `json:"browser"`
}

func NewWebhookResponseOpenedDetails(detailsRaw []byte) (WebhookResponseOpenedDetails, error) {
	var details WebhookResponseOpenedDetails
	if err := json.Unmarshal(detailsRaw, &details); err != nil {
		return WebhookResponseOpenedDetails{}, err
	}
	return details, nil
}

func (w WebhookResponseOpenedDetails) Send(response WebhookResponse) error {
	yellow := "#ffff00"
	attachment := slack.Attachment{Title: &EmailOpened, Color: &yellow}
	attachment.AddField(slack.Field{Title: "ID", Value: w.Payload.ID[0]})
	attachment.AddField(slack.Field{Title: "Email", Value: response.Email})
	attachment.AddField(slack.Field{Title: "Address", Value: w.Browser.Address})
	attachment.AddField(slack.Field{Title: "User Agent", Value: w.Browser.UserAgent})
	return sendSlack(attachment)
}

func validateSignature(body []byte, r *http.Request) error {
	secret := viper.GetString("secret")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	var actual string
	if _, err := fmt.Sscanf(r.Header.Get("X-Gophish-Signature"), "sha256=%s", &actual); err != nil {
		return errNoSignature
	}
	if !hmac.Equal([]byte(expected), []byte(actual)) {
		return errInvalidSignature
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debug(string(body))

	if err := validateSignature(body, r); err != nil {
		log.Error(errInvalidSignature)
		http.Error(w, errInvalidSignature.Error(), http.StatusBadRequest)
		return
	}

	var response WebhookResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var sender SlackSender
	if response.Message == ClickedLink {
		details, err := NewWebhookResponseClickDetails([]byte(response.Details))
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sender = details
	}
	if response.Message == SubmittedData {
		details, err := NewWebhookResponseSubmittedDetails([]byte(response.Details))
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sender = details
	}
	if response.Message == EmailOpened {
		details, err := NewWebhookResponseOpenedDetails([]byte(response.Details))
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sender = details
	}
	if err := sender.Send(response); err != nil {
		log.Error(err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/gophish_slack")
	viper.AddConfigPath(".")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("bot_username", "PhishBot")
	viper.SetDefault("bot_emoji", ":blowfish:")
	viper.SetDefault("listen_host", "0.0.0.0")
	viper.SetDefault("listen_port", "9999")
	viper.SetDefault("webhook_path", "/webhook")
	if err := viper.ReadInConfig(); err != nil {
		log.Error(err)
	}
	log.Infof("Using config file: %s", viper.ConfigFileUsed())
	vals := []string{"bot_channel", "slack_webhook", "secret"}
	for _, v := range vals {
		if !viper.IsSet(v) {
			log.Panic("Config value is not set:", v)
		}
	}
	level, err := log.ParseLevel(viper.GetString("log_level"))
	if err != nil {
		log.Panic("log level must be a valid level: panic, fatal, error, warning, info, debug, trace")
	}
	log.SetLevel(level)
}

func main() {
	addr := net.JoinHostPort(viper.GetString("listen_host"), viper.GetString("listen_port"))
	log.Infof("Server listening on %s%s", addr, viper.GetString("webhook_path"))
	http.ListenAndServe(addr, http.HandlerFunc(handler))
}

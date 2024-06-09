package main

import (
	"encoding/json"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/ashwanthkumar/slack-go-webhook"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Can't be const because need reference to variable for Slack webhook title
var (
	EmailError    string = "Error Sending Email"
	EmailSent     string = "Email Sent"
	EmailOpened   string = "Email Opened"
	ClickedLink   string = "Clicked Link"
	SubmittedData string = "Submitted Data"
	EmailReported string = "Email Reported"
)

type Sender interface {
	SendSlack() error
	SendEmail() error
	SendGraphql() error
}

func contains(slice []string, str string) bool {
	for _, a := range slice {
		if a == str {
			return true
		}
	}
	return false
}

func senderDispatch(status string, webhookResponse WebhookResponse, response []byte) (Sender, error) {
	enabled_events := viper.GetStringSlice("events")
	if status == EmailError && contains(enabled_events, "email_error") {
		return NewErrorDetails(webhookResponse)
	}
	if status == EmailSent && contains(enabled_events, "email_sent") {
		return NewSentDetails(webhookResponse)
	}
	if status == EmailOpened && contains(enabled_events, "email_opened") {
		return NewOpenedDetails(webhookResponse, response)
	}
	if status == ClickedLink && contains(enabled_events, "clicked_link") {
		return NewClickDetails(webhookResponse, response)
	}
	if status == SubmittedData && contains(enabled_events, "submitted_data") {
		return NewSubmittedDetails(webhookResponse, response)
	}
	if status == EmailReported && contains(enabled_events, "email_reported") {
		return NewReportedDetails(webhookResponse)
	}
	log.Warn("unknown status:", status)
	return nil, nil
}

// More information about events can be found here:
// https://github.com/gophish/gophish/blob/db63ee978dcd678caee0db71e5e1b91f9f293880/models/result.go#L50
type WebhookResponse struct {
	Success    bool    `json:"success"`
	CampaignID uint    `json:"campaign_id"`
	Message    string  `json:"message"`
	Details    *string `json:"details"`
	Email      string  `json:"email"`
}

func NewWebhookResponse(body []byte) (WebhookResponse, error) {
	var response WebhookResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return WebhookResponse{}, err
	}
	return response, nil
}

type EventDetails struct {
	Payload url.Values        `json:"payload"`
	Browser map[string]string `json:"browser"`
}

func NewEventDetails(detailsRaw []byte) (EventDetails, error) {
	var details EventDetails
	if err := json.Unmarshal(detailsRaw, &details); err != nil {
		return EventDetails{}, err
	}
	return details, nil
}

func (e EventDetails) ID() string {
	return e.Payload.Get("id")
}

func (e EventDetails) UserAgent() string {
	return e.Browser["user-agent"]
}
func (e EventDetails) Address() string {
	return e.Browser["address"]
}

type ErrorDetails struct {
	CampaignID uint
	Email      string
}

func NewErrorDetails(response WebhookResponse) (ErrorDetails, error) {
	errorDetails := ErrorDetails{
		CampaignID: response.CampaignID,
		Email:      response.Email,
	}
	return errorDetails, nil
}

func (w ErrorDetails) SendSlack() error {
	orange := "#ffa500"
	attachment := slack.Attachment{Title: &EmailError, Color: &orange}
	if !viper.GetBool("slack.disable_credentials") {
		attachment.AddField(slack.Field{Title: "Email", Value: w.Email})
	}
	attachment = addCampaignButton(attachment, w.CampaignID)
	return sendSlackAttachment(attachment)
}

func (w ErrorDetails) SendEmail() error {
	templateString := viper.GetString("email_send_click_template")
	body, err := getEmailBody(templateString, w)
	if err != nil {
		return err
	}
	return sendEmail("PhishBot - Error Sending Email", body)
}

func (w ErrorDetails) SendGraphql() error {
	oplog_entry := ghostwriterOplogEntry{
		UserContext: w.Email,
		Description: "Campaign ID: " + strconv.FormatUint(uint64(w.CampaignID), 10),
		Comments:    EmailError,
	}
	return sendGraphql(oplog_entry)
}

type SentDetails struct {
	CampaignID uint
	Email      string
}

func NewSentDetails(response WebhookResponse) (SentDetails, error) {
	sentDetails := SentDetails{
		CampaignID: response.CampaignID,
		Email:      response.Email,
	}
	return sentDetails, nil
}

func (w SentDetails) SendSlack() error {
	orange := "#ffa500"
	attachment := slack.Attachment{Title: &EmailSent, Color: &orange}
	if !viper.GetBool("slack.disable_credentials") {
		attachment.AddField(slack.Field{Title: "Email", Value: w.Email})
	}
	attachment = addCampaignButton(attachment, w.CampaignID)
	return sendSlackAttachment(attachment)
}

func (w SentDetails) SendEmail() error {
	templateString := viper.GetString("email_send_click_template")
	body, err := getEmailBody(templateString, w)
	if err != nil {
		return err
	}
	return sendEmail("PhishBot - Email Sent", body)
}

func (w SentDetails) SendGraphql() error {
	oplog_entry := ghostwriterOplogEntry{
		UserContext: w.Email,
		Description: "Campaign ID: " + strconv.FormatUint(uint64(w.CampaignID), 10),
		Comments:    EmailSent,
	}
	return sendGraphql(oplog_entry)
}

type OpenedDetails struct {
	CampaignID uint
	ID         string
	Email      string
	Address    string
	UserAgent  string
}

func NewOpenedDetails(response WebhookResponse, detailsRaw []byte) (OpenedDetails, error) {
	details, err := NewEventDetails(detailsRaw)
	if err != nil {
		return OpenedDetails{}, err
	}
	clickDetails := OpenedDetails{
		CampaignID: response.CampaignID,
		ID:         details.ID(),
		Email:      response.Email,
		Address:    details.Address(),
		UserAgent:  details.UserAgent(),
	}
	return clickDetails, nil
}

func (w OpenedDetails) SendSlack() error {
	yellow := "#ffff00"
	attachment := slack.Attachment{Title: &EmailOpened, Color: &yellow}
	attachment.AddField(slack.Field{Title: "ID", Value: w.ID})
	attachment.AddField(slack.Field{Title: "Address", Value: slackFormatIP(w.Address)})
	attachment.AddField(slack.Field{Title: "User Agent", Value: w.UserAgent})
	if !viper.GetBool("slack.disable_credentials") {
		attachment.AddField(slack.Field{Title: "Email", Value: w.Email})
	}
	attachment = addCampaignButton(attachment, w.CampaignID)
	return sendSlackAttachment(attachment)
}

func (w OpenedDetails) SendEmail() error {
	templateString := viper.GetString("email_send_click_template")
	body, err := getEmailBody(templateString, w)
	if err != nil {
		return err
	}
	return sendEmail("PhishBot - Email Opened", body)
}

func (w OpenedDetails) SendGraphql() error {
	oplog_entry := ghostwriterOplogEntry{
		SourceIp:    w.Address,
		UserContext: w.Email,
		Description: "User ID: " + w.ID + "\nCampaign ID: " + strconv.FormatUint(uint64(w.CampaignID), 10),
		Output:      "UserAgent: " + w.UserAgent,
		Comments:    EmailOpened,
	}
	return sendGraphql(oplog_entry)
}

type ClickDetails struct {
	CampaignID uint
	ID         string
	Email      string
	Address    string
	UserAgent  string
}

func NewClickDetails(response WebhookResponse, detailsRaw []byte) (ClickDetails, error) {
	details, err := NewEventDetails(detailsRaw)
	if err != nil {
		return ClickDetails{}, err
	}
	clickDetails := ClickDetails{
		CampaignID: response.CampaignID,
		ID:         details.ID(),
		Address:    details.Address(),
		Email:      response.Email,
		UserAgent:  details.UserAgent(),
	}
	return clickDetails, nil
}

func (w ClickDetails) SendSlack() error {
	orange := "#ffa500"
	attachment := slack.Attachment{Title: &ClickedLink, Color: &orange}
	attachment.AddField(slack.Field{Title: "ID", Value: w.ID})
	attachment.AddField(slack.Field{Title: "Address", Value: slackFormatIP(w.Address)})
	attachment.AddField(slack.Field{Title: "User Agent", Value: w.UserAgent})
	if !viper.GetBool("slack.disable_credentials") {
		attachment.AddField(slack.Field{Title: "Email", Value: w.Email})
	}
	attachment = addCampaignButton(attachment, w.CampaignID)
	return sendSlackAttachment(attachment)
}

func (w ClickDetails) SendEmail() error {
	templateString := viper.GetString("email_send_click_template")
	body, err := getEmailBody(templateString, w)
	if err != nil {
		return err
	}
	return sendEmail("PhishBot - Email Clicked", body)
}

func (w ClickDetails) SendGraphql() error {
	oplog_entry := ghostwriterOplogEntry{
		SourceIp:    w.Address,
		UserContext: w.Email,
		Description: "User ID: " + w.ID + "\nCampaign ID: " + strconv.FormatUint(uint64(w.CampaignID), 10),
		Output:      "UserAgent: " + w.UserAgent,
		Comments:    ClickedLink,
	}
	return sendGraphql(oplog_entry)
}

func getEmailBody(templateValue string, obj interface{}) (string, error) {
	out := new(strings.Builder)
	tpl, err := template.New("email").Parse(templateValue)
	if err != nil {
		return "", err
	}
	if err := tpl.Execute(out, obj); err != nil {
		return "", err
	}
	return out.String(), nil
}

type SubmittedDetails struct {
	CampaignID uint
	ID         string
	Email      string
	Address    string
	UserAgent  string
	Username   string
	Password   string
}

func NewSubmittedDetails(response WebhookResponse, detailsRaw []byte) (SubmittedDetails, error) {
	details, err := NewEventDetails(detailsRaw)
	if err != nil {
		return SubmittedDetails{}, err
	}
	submittedDetails := SubmittedDetails{
		CampaignID: response.CampaignID,
		ID:         details.ID(),
		Address:    details.Address(),
		UserAgent:  details.UserAgent(),
		Email:      response.Email,
		Username:   details.Payload.Get("username"),
		Password:   details.Payload.Get("password"),
	}
	return submittedDetails, nil
}

func (w SubmittedDetails) SendSlack() error {
	red := "#f05b4f"
	attachment := slack.Attachment{Title: &SubmittedData, Color: &red}
	attachment.AddField(slack.Field{Title: "ID", Value: w.ID})
	attachment.AddField(slack.Field{Title: "Address", Value: slackFormatIP(w.Address)})
	attachment.AddField(slack.Field{Title: "User Agent", Value: w.UserAgent})
	if !viper.GetBool("slack.disable_credentials") {
		attachment.AddField(slack.Field{Title: "Email", Value: w.Email})
		attachment.AddField(slack.Field{Title: "Username", Value: w.Username})
		attachment.AddField(slack.Field{Title: "Password", Value: w.Password})
	}
	attachment = addCampaignButton(attachment, w.CampaignID)
	return sendSlackAttachment(attachment)
}

func (w SubmittedDetails) SendEmail() error {
	templateString := viper.GetString("email_submitted_credentials_template")
	body, err := getEmailBody(templateString, w)
	if err != nil {
		return err
	}
	return sendEmail("PhishBot - Credentials Submitted", body)
}

func (w SubmittedDetails) SendGraphql() error {
	var output string
	if !viper.GetBool("ghostwriter.disable_credentials") {
		output = "\nUsername: " + w.Username + "\nPassword: " + w.Password
	}
	oplog_entry := ghostwriterOplogEntry{
		SourceIp:    w.Address,
		UserContext: w.Email,
		Description: "User ID: " + w.ID + "\nCampaign ID: " + strconv.FormatUint(uint64(w.CampaignID), 10),
		Output:      output,
		Comments:    SubmittedData,
	}
	return sendGraphql(oplog_entry)
}

type ReportedDetails struct {
	CampaignID uint
	Email      string
}

func NewReportedDetails(response WebhookResponse) (ReportedDetails, error) {
	reportedDetails := ReportedDetails{
		CampaignID: response.CampaignID,
		Email:      response.Email,
	}
	return reportedDetails, nil
}

func (w ReportedDetails) SendSlack() error {
	orange := "#ffa500"
	attachment := slack.Attachment{Title: &EmailReported, Color: &orange}
	if !viper.GetBool("slack.disable_credentials") {
		attachment.AddField(slack.Field{Title: "Email", Value: w.Email})
	}
	attachment = addCampaignButton(attachment, w.CampaignID)
	return sendSlackAttachment(attachment)
}

func (w ReportedDetails) SendEmail() error {
	templateString := viper.GetString("email_send_click_template")
	body, err := getEmailBody(templateString, w)
	if err != nil {
		return err
	}
	return sendEmail("PhishBot - Email Reported", body)
}

func (w ReportedDetails) SendGraphql() error {
	oplog_entry := ghostwriterOplogEntry{
		UserContext: w.Email,
		Description: "Campaign ID: " + strconv.FormatUint(uint64(w.CampaignID), 10),
		Comments:    EmailReported,
	}
	return sendGraphql(oplog_entry)
}

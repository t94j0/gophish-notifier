# PhishBot

GoPhish Slack bot using webhooks

## Installation

### From Source

```bash
git clone https://github.com/t94j0/gophish-webhook-slack
cd gophish-webhook-slack
go build -o phishbot
```

## Configuration

The configuration path is `/etc/gophish_slack/config.yml`. Below is an example config:

```yaml
# Username displayed in Slack
bot_username: PhishBot
# Channel to post in
bot_channel: '#testing2'
# Bot profile picture
bot_emoji: ':blowfish:'
# Host to listen on. If GoPhish is running on the same host, you can set this to 127.0.0.1
listen_host: 0.0.0.0
# Port to listen on
listen_port: 9999
# Webhook path. The full address will be http://<host>:<ip><webhook_path>. Ex: http://127.0.0.1:9999/webhook
webhook_path: /webhook
# Secret for GoPhish authentication
secret: secretpassword123
# Log level. Log levels are panic, fatal, error, warning, info, debug, trace.
log_level: debug
# Enables sending profiles. Options are `slack` and `email`. Make sure to configure the required parameters for each profile
profiles:
  - slack
  - email

# Slack Profile
# Webhook address. Typically starts with https://hooks.slack.com/services/...
slack_webhook: '<Your Slack Webhook>'

# Email Profile
# Email to send from
email_sender: test@test.com
# Password of sender email. Uses plain SMTP authentication
email_sender_password: password123
# Recipient of notifications
email_recipient: mail@example.com
# Email host to send to
email_host: smtp.gmail.com
# Email host address
email_host_addr: smtp.gmail.com:587
# You can also supply an email template for each notification
email_submitted_credentials_template: |
  Someone submitted credentials!
  Email ID - {{ .ID }}
  Email Address - {{ .Email }}
  IP Address - {{ .Address }}
  User Agent - {{ .UserAgent }}
  Username - {{ .Username }}
  Password - {{ .Password }}
```

Project inspired by [gophish-notifications]

[gophish-notifications]: https://github.com/dunderhay/gophish-notifications
# GoPhish Notifier

GoPhish Notifier notifies red team members when their GoPhish campaign status
has been updated. It supports both Slack and Email notification profiles by
default, but it's very extensible so new notification profiles can be added
easily.


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
# (Optional) Base URL of GoPhish instance so that Slack notifications can link to campaign
base_url: https://10.0.0.15:3333
# Enables sending profiles. Options are `slack` and `email`. Make sure to configure the required parameters for each profile
profiles:
  - slack
  - email

# Slack Profile
slack:
  # Webhook address. Typically starts with https://hooks.slack.com/services/...
  webhook: '<Your Slack Webhook>'
  # Username displayed in Slack
  username: PhishBot
  # Channel to post in
  channel: '#testing2'
  # Bot profile picture
  emoji: ':blowfish:'
  # (Optional) Disable email, username, and credentials from being sent to Slack
  disable_credentials: true

# Email Profile
# Email to send from
email:
  sender: test@test.com
  # Password of sender email. Uses plain SMTP authentication
  sender_password: password123
  # Recipient of notifications
  recipient: mail@example.com
  # Email host to send to
  host: smtp.gmail.com
  # Email host address
  host_addr: smtp.gmail.com:587

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

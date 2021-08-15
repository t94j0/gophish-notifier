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
# Webhook address. Typically starts with https://hooks.slack.com/services/...
slack_webhook: '<Your Slack Webhook>'
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
```

Project inspired by [gophish-notifications]

[gophish-notifications]: https://github.com/dunderhay/gophish-notifications
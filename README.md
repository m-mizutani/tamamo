# tamamo
Generative AI Agent Manager for Slack

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Slack workspace with a configured bot

### Installation

```bash
go install github.com/m-mizutani/tamamo@latest
```

Or build from source:

```bash
git clone https://github.com/m-mizutani/tamamo.git
cd tamamo
go build
```

## Running the Server

The `serve` command starts the HTTP server to handle Slack events.

### Configuration

The server requires the following configuration parameters:

| Parameter | CLI Flag | Environment Variable | Description | Required |
|-----------|----------|---------------------|-------------|----------|
| Server Address | `--addr` | `TAMAMO_ADDR` | HTTP server listen address (default: `127.0.0.1:8080`) | No |
| Slack OAuth Token | `--slack-oauth-token` | `TAMAMO_SLACK_OAUTH_TOKEN` | Bot User OAuth Token from Slack App settings | Yes |
| Slack Signing Secret | `--slack-signing-secret` | `TAMAMO_SLACK_SIGNING_SECRET` | Signing Secret for request verification from Slack App settings | Yes |

### Example Usage

Using CLI flags:

```bash
tamamo serve \
  --addr 0.0.0.0:8080 \
  --slack-oauth-token xoxb-your-token \
  --slack-signing-secret your-signing-secret
```

Using environment variables:

```bash
export TAMAMO_ADDR="0.0.0.0:8080"
export TAMAMO_SLACK_OAUTH_TOKEN="xoxb-your-token"
export TAMAMO_SLACK_SIGNING_SECRET="your-signing-secret"

tamamo serve
```

### Slack App Configuration

1. Create a new Slack App at https://api.slack.com/apps
2. Configure OAuth & Permissions:
   - Add necessary bot token scopes (e.g., `chat:write`, `app_mentions:read`)
   - Install the app to your workspace
   - Copy the Bot User OAuth Token
3. Configure Event Subscriptions:
   - Enable Events
   - Set Request URL to `http://your-server-address/hooks/slack/events`
   - Subscribe to bot events (e.g., `app_mention`, `message.channels`)
4. Copy the Signing Secret from Basic Information section

### Endpoints

The server exposes the following endpoints:

- `/hooks/slack/events` - Slack Events API webhook endpoint
- `/hooks/slack/interaction` - Slack Interactive Components endpoint (future implementation)

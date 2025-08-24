# tamamo
Generative AI Agent Manager for Slack

<p align="center">
  <img src="./docs/images/logo_v0.png" height="128" />
</p>

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
| LLM Providers Config | `--llm-providers-config` | `TAMAMO_LLM_PROVIDERS_CONFIG` | Path to LLM providers configuration file | No |

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
   - Set Request URL to `http://your-server-address/hooks/slack/event`
   - Subscribe to bot events (e.g., `app_mention`, `message.channels`)
4. Copy the Signing Secret from Basic Information section

### Endpoints

The server exposes the following endpoints:

- `/hooks/slack/events` - Slack Events API webhook endpoint
- `/hooks/slack/interaction` - Slack Interactive Components endpoint (future implementation)

## LLM Provider Configuration

Tamamo supports multiple LLM providers (OpenAI, Claude, Gemini) for agents. You can configure available providers and models using a YAML configuration file.

### Generating Configuration Template

Use the `tool generate-config` command to create a template configuration file:

```bash
tamamo tool generate-config llm-providers > providers.yaml
```

### Configuration File Format

The providers configuration file defines available LLM providers, models, and default/fallback settings:

```yaml
providers:
  openai:
    displayName: OpenAI
    models:
      - id: gpt-5-2025-08-07
        displayName: GPT-5
        description: Latest OpenAI model
      - id: gpt-5-mini-2025-08-07
        displayName: GPT-5 Mini
        description: Smaller, faster variant

  claude:
    displayName: Claude (Anthropic)
    models:
      - id: claude-sonnet-4-20250514
        displayName: Claude Sonnet 4
        description: Advanced reasoning model

  gemini:
    displayName: Google Gemini
    models:
      - id: gemini-2.5-flash
        displayName: Gemini 2.5 Flash
        description: Fast and efficient

defaults:
  provider: gemini
  model: gemini-2.0-flash

fallback:
  enabled: true
  provider: openai
  model: gpt-5-mini-2025-08-07
```

### Setting Up API Keys

Configure API keys for each provider using environment variables:

```bash
# OpenAI
export TAMAMO_OPENAI_API_KEY="sk-..."

# Claude (Anthropic)
export TAMAMO_CLAUDE_API_KEY="sk-ant-..."

# Gemini (Google Cloud)
export TAMAMO_GEMINI_PROJECT_ID="your-project-id"
export TAMAMO_GEMINI_LOCATION="us-central1"
```

### Using with the Server

Start the server with the providers configuration:

```bash
tamamo serve \
  --llm-providers-config providers.yaml \
  --slack-oauth-token xoxb-your-token \
  --slack-signing-secret your-signing-secret
```

### Agent Configuration

When creating or updating agents through the Web UI, you can:
- Select from configured LLM providers
- Choose specific models for each provider
- Agents will use their configured provider/model for processing messages
- If an agent's provider fails, the system will automatically fallback to the configured fallback provider (if enabled)

# Integration Setup Guide

This guide explains how to set up integrations with external services for Tamamo. Each integration requires service-side configuration and environment variables.

## Slack Integration

### 1. Create Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Enter app name (e.g., "Tamamo") and select your workspace
4. Click "Create App"

### 2. Configure OAuth & Permissions

1. In your app settings, go to "OAuth & Permissions"
2. Add the following Bot Token Scopes:
   - `app_mentions:read` - View messages that directly mention @your_app
   - `channels:history` - View messages and other content in public channels
   - `chat:write` - Send messages as the app
   - `files:read` - View files shared in channels and conversations
   - `groups:history` - View messages and other content in private channels
   - `im:history` - View messages and other content in direct messages
   - `mpim:history` - View messages and other content in group direct messages
   - `users:read` - View people in a workspace

3. Install the app to your workspace
4. Copy the "Bot User OAuth Token" (starts with `xoxb-`)

### 3. Configure Event Subscriptions

1. Go to "Event Subscriptions" in your app settings
2. Enable Events
3. Set Request URL to: `https://your-domain.com/hooks/slack/events`
4. Subscribe to bot events:
   - `app_mention` - A user mentions your app
   - `message.channels` - Message posted to public channel
   - `message.groups` - Message posted to private channel
   - `message.im` - Message posted to direct message channel
   - `message.mpim` - Message posted to group direct message channel

### 4. Get Signing Secret

1. Go to "Basic Information" in your app settings
2. Copy the "Signing Secret" from the "App Credentials" section

### 5. Environment Variables

Set the following environment variables:

```bash
# Slack OAuth Token (Bot User OAuth Token)
export TAMAMO_SLACK_OAUTH_TOKEN="xoxb-your-bot-token"

# Slack Signing Secret (for webhook verification)
export TAMAMO_SLACK_SIGNING_SECRET="your-signing-secret"
```

### 6. CLI Flags (Alternative)

You can also use CLI flags instead of environment variables:

```bash
tamamo serve \
  --slack-oauth-token "xoxb-your-bot-token" \
  --slack-signing-secret "your-signing-secret"
```

## JIRA Integration

### 1. Create Atlassian OAuth App

1. Go to [Atlassian Developer Console](https://developer.atlassian.com/console/myapps/)
2. Click "Create" → "OAuth 2.0 integration"
3. Enter app name (e.g., "Tamamo")
4. Select "Jira" as the product

### 2. Configure OAuth Settings

1. In your app settings, go to "Authorization"
2. Add callback URL: `https://your-domain.com/api/auth/jira/callback`

### 3. Configure API Permissions

1. Go to "Permissions" tab in your app settings
2. Add the following APIs by clicking "Add" for each:
   - **Jira API**: Add scopes for `read:jira-user`, `read:jira-work`
   - **User identity API**: Add to get basic user profile information
   - You can add other APIs later if needed for additional functionality

### 4. Complete Integration Information

Public integrations require additional information:

1. **Company Information**:
   - Company name: "Tamamo" (or your organization name)
   - Website: Your project website or GitHub repository
   - Tagline: Brief description of your integration
   - Privacy Policy URL: Link to your privacy policy
   - Terms of Use URL: Link to your terms of service
   - Email: Developer contact email
   - Logo: Upload a 512x512 pixel logo

2. **For internal use only**: You can use placeholder URLs initially
   - Website: `https://github.com/m-mizutani/tamamo`
   - Privacy Policy: `https://github.com/m-mizutani/tamamo/blob/main/PRIVACY.md`
   - Terms of Use: `https://github.com/m-mizutani/tamamo/blob/main/TERMS.md`

### 5. Get OAuth Credentials

1. In your app settings, go to "Settings"
2. Copy the "Client ID"
3. Copy the "Secret" (Client Secret)

### 6. Environment Variables

Set the following environment variables:

```bash
# JIRA OAuth Client ID
export TAMAMO_JIRA_CLIENT_ID="your-client-id"

# JIRA OAuth Client Secret
export TAMAMO_JIRA_CLIENT_SECRET="your-client-secret"

# Frontend URL for your application (used for OAuth redirects)
export TAMAMO_FRONTEND_URL="https://your-domain.com"
```

### 6. CLI Flags (Alternative)

You can also use CLI flags instead of environment variables:

```bash
tamamo serve \
  --jira-client-id "your-client-id" \
  --jira-client-secret "your-client-secret" \
  --frontend-url "https://your-domain.com"
```

## Notion Integration

### 1. Create Notion Integration

1. Go to [Notion Developers](https://www.notion.so/my-integrations)
2. Click "New integration"
3. Enter integration name (e.g., "Tamamo")
4. Select associated workspace
5. **Important**: Set Type to "Public" (required for OAuth)
   - Public integrations allow OAuth 2.0 authentication
   - Internal integrations only work with Internal Secret Tokens
   - Tamamo requires OAuth for user authentication
   - Note: While this is a Public integration, Tamamo implements additional access control to restrict usage to Slack workspace members only
6. Click "Submit"

### 2. Configure OAuth Settings

1. In your integration settings, go to "OAuth Domain & URIs"
2. Add redirect URI: `https://your-domain.com/api/auth/notion/callback`
3. Configure capabilities:
   - **Read content**: Allow reading of database and page content

### 3. Get OAuth Credentials

1. In your integration settings, copy the "OAuth client ID"
2. Copy the "OAuth client secret"

### 4. Environment Variables

Set the following environment variables:

```bash
# Notion OAuth Client ID
export TAMAMO_NOTION_CLIENT_ID="your-client-id"

# Notion OAuth Client Secret
export TAMAMO_NOTION_CLIENT_SECRET="your-client-secret"

# Frontend URL for your application (shared with JIRA)
export TAMAMO_FRONTEND_URL="https://your-domain.com"

# Allowed Notion Workspace IDs (optional - restricts access to specific workspaces)
export TAMAMO_NOTION_ALLOWED_WORKSPACES="workspace-id-1,workspace-id-2"
```

### 5. Workspace Access Control (Optional)

By default, any Notion workspace can connect to Tamamo through OAuth. To restrict access to specific workspaces, configure the allowed workspace IDs:

1. **Find Workspace ID**: There are several ways to find the Notion workspace ID:
   
   **Method 1: From Notion Web UI**
   - Open your Notion workspace in a web browser
   - Go to Settings & members (click your workspace name → Settings & members)
   - Copy the URL from your browser address bar
   - The workspace ID is in the URL: `https://www.notion.so/settings/[WORKSPACE_ID]`
   - Example: If URL is `https://www.notion.so/settings/b1234567-89ab-cdef-1234-567890abcdef`, then workspace ID is `b1234567-89ab-cdef-1234-567890abcdef`
   
   **Method 2: From any Notion page URL**
   - Open any page in your Notion workspace
   - Look at the URL: `https://www.notion.so/[WORKSPACE_ID]/Page-title-hash`
   - The workspace ID is the first part after `notion.so/`
   - Example: `https://www.notion.so/mycompany/My-Page-abc123` → workspace ID is `mycompany`
   
   **Method 3: From OAuth response (automatic)**
   - The workspace ID is automatically provided in the OAuth response when users connect through Tamamo
   - Check server logs during OAuth connection to see the workspace ID
2. **Configure Restriction**: Set the `TAMAMO_NOTION_ALLOWED_WORKSPACES` environment variable with comma-separated workspace IDs
3. **Behavior**:
   - **Without restriction**: Any Notion workspace can connect (default)
   - **With restriction**: Only specified workspace IDs are allowed to connect
   - **Access denied**: Users from non-allowed workspaces receive an error message during OAuth

**Example workspace restriction**:
```bash
# Only allow specific Notion workspaces
export TAMAMO_NOTION_ALLOWED_WORKSPACES="b1234567-89ab-cdef-1234-567890abcdef,c2345678-90ab-cdef-2345-678901abcdef"
```

### 6. CLI Flags (Alternative)

You can also use CLI flags instead of environment variables:

```bash
tamamo serve \
  --notion-client-id "your-client-id" \
  --notion-client-secret "your-client-secret" \
  --notion-allowed-workspaces "workspace-id-1,workspace-id-2" \
  --frontend-url "https://your-domain.com"
```

## Complete Configuration Example

Here's a complete example with all integrations configured:

### Environment Variables (.env file)

```bash
# Application settings
TAMAMO_FRONTEND_URL="https://your-domain.com"
TAMAMO_ADDR="0.0.0.0:8080"

# Slack Integration
TAMAMO_SLACK_OAUTH_TOKEN="xoxb-your-slack-bot-token"
TAMAMO_SLACK_SIGNING_SECRET="your-slack-signing-secret"

# JIRA Integration
TAMAMO_JIRA_CLIENT_ID="your-jira-client-id"
TAMAMO_JIRA_CLIENT_SECRET="your-jira-client-secret"

# Notion Integration
TAMAMO_NOTION_CLIENT_ID="your-notion-client-id"
TAMAMO_NOTION_CLIENT_SECRET="your-notion-client-secret"
TAMAMO_NOTION_ALLOWED_WORKSPACES="workspace-id-1,workspace-id-2"

# Authentication (optional)
TAMAMO_AUTH_CLIENT_ID="your-auth-client-id"
TAMAMO_AUTH_CLIENT_SECRET="your-auth-client-secret"

# Database (optional - uses in-memory if not configured)
TAMAMO_FIRESTORE_PROJECT_ID="your-firestore-project"
TAMAMO_FIRESTORE_DATABASE_ID="(default)"
```

### CLI Command

```bash
tamamo serve \
  --addr "0.0.0.0:8080" \
  --frontend-url "https://your-domain.com" \
  --slack-oauth-token "xoxb-your-slack-bot-token" \
  --slack-signing-secret "your-slack-signing-secret" \
  --jira-client-id "your-jira-client-id" \
  --jira-client-secret "your-jira-client-secret" \
  --notion-client-id "your-notion-client-id" \
  --notion-client-secret "your-notion-client-secret" \
  --notion-allowed-workspaces "workspace-id-1,workspace-id-2"
```

## Verifying Integration Setup

### Check Configuration

Use the following command to verify your configuration:

```bash
tamamo serve --help
```

This will show all available configuration options.

### Test Slack Integration

1. Start the server with Slack configuration
2. Mention your bot in a Slack channel: `@tamamo hello`
3. Check server logs for incoming webhook events

### Test OAuth Integrations

1. Start the server with OAuth configurations
2. Access the web interface at your base URL
3. Go to Settings → Integrations
4. Try connecting to JIRA and Notion
5. Verify successful OAuth flow completion

## Troubleshooting

### Common Issues

1. **Slack events not received**
   - Verify the webhook URL is accessible from the internet
   - Check that the signing secret is correct
   - Ensure bot has necessary permissions

2. **OAuth redirect mismatch**
   - Verify redirect URIs in service settings match your configuration
   - Ensure HTTPS is used for production deployments
   - Check that base URL is correctly configured

3. **Permission denied errors**
   - Review and update OAuth scopes/permissions
   - Re-authorize the integration if permissions changed

### Logs

Enable debug logging to troubleshoot issues:

```bash
tamamo serve --log-level debug
```

This will provide detailed information about webhook processing and OAuth flows.

## Access Control

### Integration Access Control

Tamamo implements different access control mechanisms for each integration:

#### Jira Integration Access Control
1. **User Authentication**: Users must first authenticate through Slack to access Tamamo
2. **Slack Workspace Verification**: When connecting Jira, Tamamo verifies that the user's email is associated with a member of the Slack workspace
3. **Real-time Validation**: The verification happens during the OAuth callback process
4. **Error Handling**: Non-workspace members receive a clear error message: "Access denied: You must be a member of the Slack workspace to connect Jira."

#### Notion Integration Access Control
1. **User Authentication**: Users must first authenticate through Slack to access Tamamo
2. **Notion Workspace Restriction**: Access is controlled by allowed Notion workspace IDs (configured via `TAMAMO_NOTION_ALLOWED_WORKSPACES`)
3. **Behavior**:
   - **Default (no restriction)**: Any Notion workspace can connect through OAuth
   - **With restriction**: Only specified Notion workspace IDs are allowed to connect
   - **Access denied**: Users from non-allowed workspaces receive an error message: "Access denied: This Notion workspace is not allowed to connect to Tamamo."
4. **Real-time Validation**: The workspace validation happens during the OAuth callback process using the workspace ID from Notion's OAuth response

This ensures that even though Jira and Notion integrations are configured as "Public" (required for OAuth), access is effectively restricted based on your access control policies.

### Benefits

- **Flexible Access Control**: Different access control mechanisms for different integrations
- **Individual Permissions**: Each user connects with their own credentials and permissions
- **Centralized Control**: 
  - **Jira**: Controlled through Slack workspace membership
  - **Notion**: Controlled through allowed Notion workspace IDs
- **Audit Trail**: All integration activities are tied to authenticated Slack users
- **Granular Restrictions**: Fine-grained control over which workspaces can connect

## Security Considerations

1. **Use HTTPS in production** - OAuth requires secure connections
2. **Keep secrets secure** - Never commit OAuth credentials to version control
3. **Rotate secrets regularly** - Update OAuth credentials periodically
4. **Verify webhook signatures** - Tamamo automatically verifies Slack webhook signatures
5. **Use environment variables** - Avoid hardcoding credentials in configuration files
6. **Integration access control** - Configure appropriate access restrictions for each integration:
   - **Jira**: Automatically restricted to Slack workspace members
   - **Notion**: Configure `TAMAMO_NOTION_ALLOWED_WORKSPACES` to restrict access to specific Notion workspaces

## Next Steps

Once integrations are configured:

1. Configure authentication if using a public deployment
2. Set up persistent storage (Firestore) for production use
3. Configure LLM providers for AI functionality
4. Review and adjust OAuth scopes based on your use case
import { useState, useEffect, useRef } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ConfirmDialog } from '@/components/ConfirmDialog'
import { executeGraphQL, GET_JIRA_INTEGRATION, INITIATE_JIRA_OAUTH, DISCONNECT_JIRA } from '@/lib/graphql'
import { toast } from 'sonner'
import { ExternalLink, Link, Unlink, RefreshCw } from 'lucide-react'

interface JiraIntegration {
  id: string
  connected: boolean
  siteUrl?: string
  connectedAt?: string
}

export function JiraIntegrationCard() {
  const [integration, setIntegration] = useState<JiraIntegration | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState(false)
  const [showDisconnectDialog, setShowDisconnectDialog] = useState(false)
  const popupPollInterval = useRef<number | null>(null)

  const loadIntegration = async () => {
    try {
      setLoading(true)
      const data = await executeGraphQL(GET_JIRA_INTEGRATION)
      setIntegration(data.jiraIntegration)
    } catch (error) {
      console.error('Failed to load Jira integration:', error)
      toast.error('Failed to load Jira integration status')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadIntegration()
    
    // Cleanup function to clear interval on unmount
    return () => {
      if (popupPollInterval.current) {
        clearInterval(popupPollInterval.current)
        popupPollInterval.current = null
      }
    }
  }, [])

  const handleConnect = async () => {
    try {
      setActionLoading(true)
      const data = await executeGraphQL(INITIATE_JIRA_OAUTH, {})
      
      if (data.initiateJiraOAuth?.url) {
        // Open OAuth URL in new window
        const width = 600
        const height = 700
        const left = window.screenX + (window.outerWidth - width) / 2
        const top = window.screenY + (window.outerHeight - height) / 2.5
        
        const popup = window.open(
          data.initiateJiraOAuth.url,
          'jira-oauth',
          `width=${width},height=${height},left=${left},top=${top},scrollbars=yes,resizable=yes`
        )

        if (popup) {
          // Clear any existing interval before starting a new one
          if (popupPollInterval.current) {
            clearInterval(popupPollInterval.current)
          }

          // Poll for popup closure to refresh status
          popupPollInterval.current = window.setInterval(() => {
            if (popup.closed) {
              if (popupPollInterval.current) {
                clearInterval(popupPollInterval.current)
                popupPollInterval.current = null
              }
              // Refresh integration status after popup closes
              setTimeout(() => {
                loadIntegration()
              }, 1000)
            }
          }, 1000)
        } else {
          toast.error('Failed to open authentication window. Please allow popups.')
        }
      }
    } catch (error) {
      console.error('Failed to initiate OAuth:', error)
      toast.error('Failed to start Jira connection process')
    } finally {
      setActionLoading(false)
    }
  }

  const handleDisconnect = async () => {
    try {
      setActionLoading(true)
      await executeGraphQL(DISCONNECT_JIRA, {})
      toast.success('Successfully disconnected from Jira')
      setShowDisconnectDialog(false)
      await loadIntegration()
    } catch (error) {
      console.error('Failed to disconnect:', error)
      toast.error('Failed to disconnect from Jira')
    } finally {
      setActionLoading(false)
    }
  }

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-blue-100 rounded flex items-center justify-center">
              <RefreshCw className="h-4 w-4 text-blue-600 animate-spin" />
            </div>
            <div>
              <CardTitle className="text-lg">Jira</CardTitle>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Loading integration status...</p>
        </CardContent>
      </Card>
    )
  }

  const isConnected = integration?.connected ?? false
  const siteUrl = integration?.siteUrl
  const connectedAt = integration?.connectedAt

  return (
    <>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-blue-100 rounded flex items-center justify-center">
                {isConnected ? (
                  <Link className="h-4 w-4 text-blue-600" />
                ) : (
                  <Unlink className="h-4 w-4 text-gray-400" />
                )}
              </div>
              <div>
                <CardTitle className="text-lg">Jira</CardTitle>
                <p className="text-sm text-muted-foreground">
                  Issue tracking and project management
                </p>
              </div>
            </div>
            <Badge variant={isConnected ? "default" : "secondary"}>
              {isConnected ? "Connected" : "Not connected"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {isConnected ? (
            <div className="space-y-3">
              {siteUrl && (
                <div className="flex items-center justify-between p-3 bg-muted/50 rounded-lg">
                  <div>
                    <p className="text-sm font-medium">Connected to:</p>
                    <p className="text-sm text-muted-foreground">{siteUrl}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => window.open(`https://${siteUrl}`, '_blank')}
                  >
                    <ExternalLink className="h-4 w-4" />
                  </Button>
                </div>
              )}
              {connectedAt && (
                <p className="text-xs text-muted-foreground">
                  Connected on {new Date(connectedAt).toLocaleDateString()}
                </p>
              )}
              <div className="flex space-x-2">
                <Button
                  variant="outline"
                  onClick={() => setShowDisconnectDialog(true)}
                  disabled={actionLoading}
                >
                  Disconnect
                </Button>
              </div>
            </div>
          ) : (
            <div className="space-y-3">
              <p className="text-sm text-muted-foreground">
                Connect your Jira account to enable issue tracking integration with Tamamo.
              </p>
              <Button 
                onClick={handleConnect}
                disabled={actionLoading}
                className="w-full"
              >
                {actionLoading ? (
                  <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                ) : (
                  <Link className="h-4 w-4 mr-2" />
                )}
                Connect to Jira
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={showDisconnectDialog}
        onOpenChange={setShowDisconnectDialog}
        onConfirm={handleDisconnect}
        title="Disconnect from Jira"
        description="Are you sure you want to disconnect from Jira? This will remove access to your Jira account and any related integrations will stop working."
        confirmText="Disconnect"
        confirmVariant="destructive"
        loading={actionLoading}
      />
    </>
  )
}
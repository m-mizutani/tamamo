import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { X, Save } from 'lucide-react'
import {
  AgentSlackSearchConfig,
  CreateSlackSearchConfigInput,
  UpdateSlackSearchConfigInput
} from '@/lib/graphql'

interface Props {
  config?: AgentSlackSearchConfig | null
  onSubmit: (data: Omit<CreateSlackSearchConfigInput, 'agentId'> | UpdateSlackSearchConfigInput) => Promise<void>
  onCancel: () => void
}

export function SlackSearchConfigForm({ config, onSubmit, onCancel }: Props) {
  const [channelId, setChannelId] = useState(config?.channelId || '')
  const [channelName, setChannelName] = useState(config?.channelName || '')
  const [description, setDescription] = useState(config?.description || '')
  const [enabled, setEnabled] = useState(config?.enabled ?? true)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!channelName.trim() || (!config && !channelId.trim())) {
      return
    }

    setSubmitting(true)
    try {
      if (config) {
        // Update existing config
        await onSubmit({
          channelName: channelName.trim(),
          description: description.trim() || undefined,
          enabled
        } as UpdateSlackSearchConfigInput)
      } else {
        // Create new config
        await onSubmit({
          channelId: channelId.trim(),
          channelName: channelName.trim(),
          description: description.trim() || undefined,
          enabled
        } as Omit<CreateSlackSearchConfigInput, 'agentId'>)
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <form onSubmit={handleSubmit} className="space-y-4">
          {!config && (
            <div className="space-y-2">
              <Label htmlFor="channelId">Channel ID</Label>
              <Input
                id="channelId"
                placeholder="C1234567890"
                value={channelId}
                onChange={(e) => setChannelId(e.target.value)}
                required
                disabled={submitting}
              />
              <p className="text-xs text-muted-foreground">
                The Slack channel ID (starts with C)
              </p>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="channelName">Channel Name</Label>
            <Input
              id="channelName"
              placeholder="#general"
              value={channelName}
              onChange={(e) => setChannelName(e.target.value)}
              required
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              The display name of the channel
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description (Optional)</Label>
            <Textarea
              id="description"
              placeholder="Main discussion channel for the team"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={submitting}
              rows={3}
            />
          </div>

          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="enabled"
              checked={enabled}
              onChange={(e) => setEnabled(e.target.checked)}
              disabled={submitting}
              className="h-4 w-4"
            />
            <Label htmlFor="enabled">Enable search for this channel</Label>
          </div>

          <div className="flex justify-end space-x-2">
            <Button
              type="button"
              variant="outline"
              onClick={onCancel}
              disabled={submitting}
            >
              <X className="h-4 w-4 mr-2" />
              Cancel
            </Button>
            <Button type="submit" disabled={submitting}>
              <Save className="h-4 w-4 mr-2" />
              {config ? 'Update' : 'Create'}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
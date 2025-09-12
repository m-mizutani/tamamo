import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { X, Save } from 'lucide-react'
import {
  AgentNotionSearchConfig,
  CreateNotionSearchConfigInput,
  UpdateNotionSearchConfigInput
} from '@/lib/graphql'

interface Props {
  config?: AgentNotionSearchConfig | null
  onSubmit: (data: Omit<CreateNotionSearchConfigInput, 'agentId'> | UpdateNotionSearchConfigInput) => Promise<void>
  onCancel: () => void
}

export function NotionSearchConfigForm({ config, onSubmit, onCancel }: Props) {
  const [databaseId, setDatabaseId] = useState(config?.databaseId || '')
  const [databaseName, setDatabaseName] = useState(config?.databaseName || '')
  const [workspaceId, setWorkspaceId] = useState(config?.workspaceId || '')
  const [description, setDescription] = useState(config?.description || '')
  const [enabled, setEnabled] = useState(config?.enabled ?? true)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!databaseName.trim() || !workspaceId.trim() || (!config && !databaseId.trim())) {
      return
    }

    setSubmitting(true)
    try {
      if (config) {
        // Update existing config
        await onSubmit({
          databaseName: databaseName.trim(),
          workspaceId: workspaceId.trim(),
          description: description.trim() || undefined,
          enabled
        } as UpdateNotionSearchConfigInput)
      } else {
        // Create new config
        await onSubmit({
          databaseId: databaseId.trim(),
          databaseName: databaseName.trim(),
          workspaceId: workspaceId.trim(),
          description: description.trim() || undefined,
          enabled
        } as Omit<CreateNotionSearchConfigInput, 'agentId'>)
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
              <Label htmlFor="databaseId">Database ID</Label>
              <Input
                id="databaseId"
                placeholder="32-character hexadecimal string"
                value={databaseId}
                onChange={(e) => setDatabaseId(e.target.value)}
                required
                disabled={submitting}
              />
              <p className="text-xs text-muted-foreground">
                The Notion database ID (found in the database URL)
              </p>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="databaseName">Database Name</Label>
            <Input
              id="databaseName"
              placeholder="Project Database"
              value={databaseName}
              onChange={(e) => setDatabaseName(e.target.value)}
              required
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              The display name of the Notion database
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="workspaceId">Workspace ID</Label>
            <Input
              id="workspaceId"
              placeholder="workspace-id-string"
              value={workspaceId}
              onChange={(e) => setWorkspaceId(e.target.value)}
              required
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              The Notion workspace identifier
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description (Optional)</Label>
            <Textarea
              id="description"
              placeholder="Main project database for tracking tasks"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={submitting}
              rows={3}
            />
          </div>

          <div className="flex items-center space-x-2">
            <Checkbox
              id="enabled"
              checked={enabled}
              onCheckedChange={(checked) => setEnabled(Boolean(checked))}
              disabled={submitting}
            />
            <Label htmlFor="enabled">Enable search for this database</Label>
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
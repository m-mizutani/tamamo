import { useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { X, Save } from 'lucide-react'
import {
  AgentJiraSearchConfig,
  CreateJiraSearchConfigInput,
  UpdateJiraSearchConfigInput
} from '@/lib/graphql'

interface Props {
  config?: AgentJiraSearchConfig | null
  onSubmit: (data: Omit<CreateJiraSearchConfigInput, 'agentId'> | UpdateJiraSearchConfigInput) => Promise<void>
  onCancel: () => void
}

export function JiraSearchConfigForm({ config, onSubmit, onCancel }: Props) {
  const [projectKey, setProjectKey] = useState(config?.projectKey || '')
  const [projectName, setProjectName] = useState(config?.projectName || '')
  const [boardId, setBoardId] = useState(config?.boardId || '')
  const [boardName, setBoardName] = useState(config?.boardName || '')
  const [description, setDescription] = useState(config?.description || '')
  const [enabled, setEnabled] = useState(config?.enabled ?? true)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!projectName.trim() || (!config && !projectKey.trim())) {
      return
    }

    setSubmitting(true)
    try {
      if (config) {
        // Update existing config
        await onSubmit({
          projectName: projectName.trim(),
          boardId: boardId.trim() || undefined,
          boardName: boardName.trim() || undefined,
          description: description.trim() || undefined,
          enabled
        } as UpdateJiraSearchConfigInput)
      } else {
        // Create new config
        await onSubmit({
          projectKey: projectKey.trim(),
          projectName: projectName.trim(),
          boardId: boardId.trim() || undefined,
          boardName: boardName.trim() || undefined,
          description: description.trim() || undefined,
          enabled
        } as Omit<CreateJiraSearchConfigInput, 'agentId'>)
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
              <Label htmlFor="projectKey">Project Key</Label>
              <Input
                id="projectKey"
                placeholder="PROJ"
                value={projectKey}
                onChange={(e) => setProjectKey(e.target.value)}
                required
                disabled={submitting}
              />
              <p className="text-xs text-muted-foreground">
                The Jira project key (e.g., PROJ, DEV, TICKET)
              </p>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="projectName">Project Name</Label>
            <Input
              id="projectName"
              placeholder="My Project"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              required
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              The display name of the Jira project
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="boardId">Board ID (Optional)</Label>
            <Input
              id="boardId"
              placeholder="123"
              value={boardId}
              onChange={(e) => setBoardId(e.target.value)}
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              Specific board ID to search within the project
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="boardName">Board Name (Optional)</Label>
            <Input
              id="boardName"
              placeholder="Sprint Board"
              value={boardName}
              onChange={(e) => setBoardName(e.target.value)}
              disabled={submitting}
            />
            <p className="text-xs text-muted-foreground">
              Display name of the board (if board ID is specified)
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description (Optional)</Label>
            <Textarea
              id="description"
              placeholder="Main project for feature development"
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
            <Label htmlFor="enabled">Enable search for this project</Label>
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
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Edit, Trash2, Box } from 'lucide-react'
import { AgentJiraSearchConfig } from '@/lib/graphql'
import { ConfirmDialog } from '@/components/ConfirmDialog'

interface Props {
  configs: AgentJiraSearchConfig[]
  loading: boolean
  canEdit: boolean
  onEdit: (config: AgentJiraSearchConfig) => void
  onDelete: (id: string) => void
}

export function JiraSearchConfigList({ configs, loading, canEdit, onEdit, onDelete }: Props) {
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null)

  if (loading) {
    return (
      <div className="space-y-2">
        {[1, 2, 3].map((i) => (
          <div key={i} className="p-4 border rounded-lg">
            <div className="h-4 w-32 mb-2 bg-gray-200 animate-pulse rounded" />
            <div className="h-3 w-48 bg-gray-200 animate-pulse rounded" />
          </div>
        ))}
      </div>
    )
  }

  if (configs.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <Box className="h-12 w-12 mx-auto mb-3 opacity-50" />
        <p>No Jira projects configured</p>
        <p className="text-sm mt-1">Add projects to enable search functionality</p>
      </div>
    )
  }

  return (
    <>
      <div className="space-y-2">
        {configs.map((config) => (
          <div
            key={config.id}
            className="p-4 border rounded-lg hover:bg-muted/50 transition-colors"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <Box className="h-4 w-4 text-muted-foreground" />
                  <span className="font-medium">{config.projectName}</span>
                  <Badge variant="outline" className="text-xs">
                    {config.projectKey}
                  </Badge>
                  <Badge variant={config.enabled ? 'default' : 'secondary'}>
                    {config.enabled ? 'Enabled' : 'Disabled'}
                  </Badge>
                </div>
                {config.boardName && (
                  <p className="text-sm text-muted-foreground">
                    Board: {config.boardName} {config.boardId && `(${config.boardId})`}
                  </p>
                )}
                {config.description && (
                  <p className="text-sm mt-1">{config.description}</p>
                )}
                <p className="text-xs text-muted-foreground mt-2">
                  Created: {new Date(config.createdAt).toLocaleDateString()}
                </p>
              </div>
              {canEdit && (
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onEdit(config)}
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setDeleteConfirmId(config.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>

      <ConfirmDialog
        open={deleteConfirmId !== null}
        onOpenChange={(open) => !open && setDeleteConfirmId(null)}
        onConfirm={() => {
          if (deleteConfirmId) {
            onDelete(deleteConfirmId)
            setDeleteConfirmId(null)
          }
        }}
        title="Delete Search Configuration"
        description="Are you sure you want to delete this Jira project search configuration? This action cannot be undone."
        confirmText="Delete"
        confirmVariant="destructive"
      />
    </>
  )
}
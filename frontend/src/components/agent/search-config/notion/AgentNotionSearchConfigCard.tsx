import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'
import { NotionSearchConfigList } from './NotionSearchConfigList'
import { NotionSearchConfigForm } from './NotionSearchConfigForm'
import { useNotionSearchConfigs } from '@/hooks/useNotionSearchConfigs'
import {
  AgentNotionSearchConfig,
  CreateNotionSearchConfigInput,
  UpdateNotionSearchConfigInput
} from '@/lib/graphql'

interface Props {
  agentId: string
  canEdit: boolean
}

export function AgentNotionSearchConfigCard({ agentId, canEdit }: Props) {
  const [showForm, setShowForm] = useState(false)
  const [editingConfig, setEditingConfig] = useState<AgentNotionSearchConfig | null>(null)
  const { configs, loading, loadConfigs, createConfig, updateConfig, deleteConfig } = useNotionSearchConfigs(agentId)

  useEffect(() => {
    loadConfigs()
  }, [agentId, loadConfigs])

  const handleCreate = async (data: Omit<CreateNotionSearchConfigInput, 'agentId'>) => {
    await createConfig({ ...data, agentId })
    setShowForm(false)
  }

  const handleUpdate = async (id: string, data: UpdateNotionSearchConfigInput) => {
    await updateConfig(id, data)
    setEditingConfig(null)
  }

  const handleDelete = async (id: string) => {
    await deleteConfig(id)
  }

  const onFormSubmit = async (data: Omit<CreateNotionSearchConfigInput, 'agentId'> | UpdateNotionSearchConfigInput) => {
    if (editingConfig) {
      await handleUpdate(editingConfig.id, data as UpdateNotionSearchConfigInput)
    } else {
      await handleCreate(data as Omit<CreateNotionSearchConfigInput, 'agentId'>)
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Notion Database Search Config</CardTitle>
            <p className="text-sm text-muted-foreground mt-1">
              Configure Notion databases for search targets
            </p>
          </div>
          {canEdit && (
            <Button onClick={() => setShowForm(true)} size="sm">
              <Plus className="h-4 w-4 mr-2" />
              Add
            </Button>
          )}
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        {(showForm || editingConfig) && (
          <NotionSearchConfigForm
            config={editingConfig}
            onSubmit={onFormSubmit}
            onCancel={() => {
              setShowForm(false)
              setEditingConfig(null)
            }}
          />
        )}
        
        <NotionSearchConfigList
          configs={configs}
          loading={loading}
          canEdit={canEdit}
          onEdit={setEditingConfig}
          onDelete={handleDelete}
        />
      </CardContent>
    </Card>
  )
}
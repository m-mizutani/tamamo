import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'
import { SlackSearchConfigList } from './SlackSearchConfigList'
import { SlackSearchConfigForm } from './SlackSearchConfigForm'
import { useSlackSearchConfigs } from '@/hooks/useSlackSearchConfigs'
import {
  AgentSlackSearchConfig,
  CreateSlackSearchConfigInput,
  UpdateSlackSearchConfigInput
} from '@/lib/graphql'

interface Props {
  agentId: string
  canEdit: boolean
}

export function AgentSlackSearchConfigCard({ agentId, canEdit }: Props) {
  const [showForm, setShowForm] = useState(false)
  const [editingConfig, setEditingConfig] = useState<AgentSlackSearchConfig | null>(null)
  const { configs, loading, loadConfigs, createConfig, updateConfig, deleteConfig } = useSlackSearchConfigs(agentId)

  useEffect(() => {
    loadConfigs()
  }, [agentId, loadConfigs])

  const handleCreate = async (data: Omit<CreateSlackSearchConfigInput, 'agentId'>) => {
    await createConfig({ ...data, agentId })
    setShowForm(false)
  }

  const handleUpdate = async (id: string, data: UpdateSlackSearchConfigInput) => {
    await updateConfig(id, data)
    setEditingConfig(null)
  }

  const handleDelete = async (id: string) => {
    await deleteConfig(id)
  }

  const onFormSubmit = async (data: Omit<CreateSlackSearchConfigInput, 'agentId'> | UpdateSlackSearchConfigInput) => {
    if (editingConfig) {
      await handleUpdate(editingConfig.id, data as UpdateSlackSearchConfigInput)
    } else {
      await handleCreate(data as Omit<CreateSlackSearchConfigInput, 'agentId'>)
    }
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Slack Channel Search Config</CardTitle>
            <p className="text-sm text-muted-foreground mt-1">
              Configure Slack channels for search targets
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
          <SlackSearchConfigForm
            config={editingConfig}
            onSubmit={onFormSubmit}
            onCancel={() => {
              setShowForm(false)
              setEditingConfig(null)
            }}
          />
        )}
        
        <SlackSearchConfigList
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
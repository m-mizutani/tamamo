import { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'
import { JiraSearchConfigList } from './JiraSearchConfigList'
import { JiraSearchConfigForm } from './JiraSearchConfigForm'
import { useJiraSearchConfigs } from '@/hooks/useJiraSearchConfigs'
import {
  AgentJiraSearchConfig,
  CreateJiraSearchConfigInput,
  UpdateJiraSearchConfigInput
} from '@/lib/graphql'

interface Props {
  agentId: string
  canEdit: boolean
}

export function AgentJiraSearchConfigCard({ agentId, canEdit }: Props) {
  const [showForm, setShowForm] = useState(false)
  const [editingConfig, setEditingConfig] = useState<AgentJiraSearchConfig | null>(null)
  const { configs, loading, loadConfigs, createConfig, updateConfig, deleteConfig } = useJiraSearchConfigs(agentId)

  useEffect(() => {
    loadConfigs()
  }, [agentId, loadConfigs])

  const handleCreate = async (data: Omit<CreateJiraSearchConfigInput, 'agentId'>) => {
    await createConfig({ ...data, agentId })
    setShowForm(false)
  }

  const handleUpdate = async (id: string, data: UpdateJiraSearchConfigInput) => {
    await updateConfig(id, data)
    setEditingConfig(null)
  }

  const handleDelete = async (id: string) => {
    await deleteConfig(id)
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="text-lg">Jira Project Search Config</CardTitle>
            <p className="text-sm text-muted-foreground mt-1">
              Configure Jira projects for search targets
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
          <JiraSearchConfigForm
            config={editingConfig}
            onSubmit={editingConfig ? 
              (data) => handleUpdate(editingConfig.id, data as UpdateJiraSearchConfigInput) : 
              (data) => handleCreate(data as Omit<CreateJiraSearchConfigInput, 'agentId'>)
            }
            onCancel={() => {
              setShowForm(false)
              setEditingConfig(null)
            }}
          />
        )}
        
        <JiraSearchConfigList
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